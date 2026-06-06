package installer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	platformCACert         = "/var/snap/platform/current/syncloud.ca.crt"
	routerBasePath         = "/" + App
	oidcWebCallbackPath    = routerBasePath + "/auth/openid/callback"
	oidcMobileRedirectPath = routerBasePath + "/auth/openid/mobile-redirect"
	serverSettingsKey      = "server-settings"
	adminUsername          = "admin"
)

type absStatus struct {
	IsInit bool `json:"isInit"`
}

type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

type Oidc struct {
	platformClient *platform.Client
	logger         *zap.Logger
	dataDir        string
	client         *http.Client
}

func NewOidc(platformClient *platform.Client, logger *zap.Logger, dataDir string) *Oidc {
	return &Oidc{
		platformClient: platformClient,
		logger:         logger,
		dataDir:        dataDir,
		client:         socketHTTPClient(path.Join(dataDir, "audiobookshelf.sock")),
	}
}

func (o *Oidc) Initialize() error {
	isInit, err := o.waitForStatus()
	if err != nil {
		return err
	}
	if isInit {
		return nil
	}
	return o.createRootUser()
}

func (o *Oidc) ConfigureApp(storageDir string) error {
	dbPath := path.Join(storageDir, "config", "absdatabase.sqlite")
	secret, err := o.platformClient.RegisterOIDCClient(App, []string{oidcWebCallbackPath, oidcMobileRedirectPath}, true, "client_secret_basic")
	if err != nil {
		return fmt.Errorf("oidc register: %w", err)
	}
	authUrl, err := o.platformClient.GetAppUrl("auth")
	if err != nil {
		return err
	}
	discovery := o.discoverOIDC(authUrl)
	if err := o.enableOIDCInDb(dbPath, discovery, secret); err != nil {
		return fmt.Errorf("oidc update db: %w", err)
	}
	return o.platformClient.RestartService(fmt.Sprint(App, ".abs"))
}

func socketHTTPClient(socket string) *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socket)
			},
		},
	}
}

func (o *Oidc) waitForStatus() (bool, error) {
	for attempt := 0; attempt < 60; attempt++ {
		resp, err := o.client.Get("http://localhost/status")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var status absStatus
				if json.Unmarshal(body, &status) == nil {
					return status.IsInit, nil
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return false, fmt.Errorf("audiobookshelf did not become ready")
}

func (o *Oidc) createRootUser() error {
	password, err := randomPassword()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]interface{}{
		"newRoot": map[string]string{"username": adminUsername, "password": password},
	})
	if err != nil {
		return err
	}
	resp, err := o.client.Post("http://localhost/init", "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("init returned %s", resp.Status)
	}
	passwordFile := path.Join(o.dataDir, "initial_admin_password")
	if err := os.WriteFile(passwordFile, []byte(password+"\n"), 0600); err != nil {
		return err
	}
	o.logger.Info("created root admin user", zap.String("username", adminUsername), zap.String("password_file", passwordFile))
	return nil
}

func randomPassword() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (o *Oidc) discoverOIDC(authUrl string) *oidcDiscovery {
	base := strings.TrimRight(authUrl, "/")
	fallback := &oidcDiscovery{
		Issuer:                base,
		AuthorizationEndpoint: base + "/api/oidc/authorization",
		TokenEndpoint:         base + "/api/oidc/token",
		UserinfoEndpoint:      base + "/api/oidc/userinfo",
		JwksURI:               base + "/jwks.json",
		EndSessionEndpoint:    base + "/api/oidc/end-session",
	}

	transport := &http.Transport{}
	if caCert, err := os.ReadFile(platformCACert); err == nil {
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM(caCert) {
			transport.TLSClientConfig = &tls.Config{RootCAs: pool}
		}
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: transport}
	url := base + "/.well-known/openid-configuration"

	for attempt := 0; attempt < 30; attempt++ {
		resp, err := client.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				discovery := &oidcDiscovery{}
				if json.Unmarshal(body, discovery) == nil && discovery.TokenEndpoint != "" {
					return discovery
				}
			}
		} else {
			o.logger.Info("oidc discovery retry", zap.Error(err))
		}
		time.Sleep(2 * time.Second)
	}
	o.logger.Info("oidc discovery unavailable, using authelia defaults", zap.String("issuer", base))
	return fallback
}

func (o *Oidc) enableOIDCInDb(dbPath string, discovery *oidcDiscovery, secret string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var raw string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, serverSettingsKey).Scan(&raw); err != nil {
		return err
	}
	settings := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return err
	}

	settings["authActiveAuthMethods"] = []string{"openid"}
	settings["authOpenIDSubfolderForRedirectURLs"] = routerBasePath
	settings["authOpenIDGroupClaim"] = "groups"
	settings["authOpenIDAdminGroups"] = "syncloud"
	settings["authOpenIDGroupDefaultRole"] = "user"
	settings["authOpenIDIssuerURL"] = discovery.Issuer
	settings["authOpenIDAuthorizationURL"] = discovery.AuthorizationEndpoint
	settings["authOpenIDTokenURL"] = discovery.TokenEndpoint
	settings["authOpenIDUserInfoURL"] = discovery.UserinfoEndpoint
	settings["authOpenIDJwksURL"] = discovery.JwksURI
	if discovery.EndSessionEndpoint != "" {
		settings["authOpenIDLogoutURL"] = discovery.EndSessionEndpoint
	}
	settings["authOpenIDClientID"] = App
	settings["authOpenIDClientSecret"] = secret
	settings["authOpenIDTokenSigningAlgorithm"] = "RS256"
	settings["authOpenIDButtonText"] = "Login with Syncloud"
	settings["authOpenIDAutoRegister"] = true

	out, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE settings SET value = ? WHERE key = ?`, string(out), serverSettingsKey)
	return err
}
