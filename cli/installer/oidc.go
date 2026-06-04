package installer

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	platformCACert    = "/var/snap/platform/current/syncloud.ca.crt"
	oidcCallbackPath  = "/auth/openid/callback"
	serverSettingsKey = "server-settings"
)

type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

func (i *Installer) ConfigureOIDC(storageDir string) error {
	dbPath := path.Join(storageDir, "config", "absdatabase.sqlite")
	if err := i.waitForServerSettings(dbPath); err != nil {
		return err
	}
	secret, err := i.platformClient.RegisterOIDCClient(App, oidcCallbackPath, true, "client_secret_basic")
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	authUrl, err := i.platformClient.GetAppUrl("auth")
	if err != nil {
		return err
	}
	discovery := i.discoverOIDC(authUrl)
	if err := i.enableOIDCInDb(dbPath, discovery, secret); err != nil {
		return fmt.Errorf("update db: %w", err)
	}
	return i.platformClient.RestartService(fmt.Sprint(App, ".abs"))
}

func (i *Installer) waitForServerSettings(dbPath string) error {
	for attempt := 0; attempt < 60; attempt++ {
		if _, err := os.Stat(dbPath); err == nil {
			db, err := sql.Open("sqlite", dbPath)
			if err == nil {
				var value string
				queryErr := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, serverSettingsKey).Scan(&value)
				db.Close()
				if queryErr == nil {
					return nil
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("server-settings not found in %s", dbPath)
}

func (i *Installer) discoverOIDC(authUrl string) *oidcDiscovery {
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
			i.logger.Info("oidc discovery retry", zap.Error(err))
		}
		time.Sleep(2 * time.Second)
	}
	i.logger.Info("oidc discovery unavailable, using authelia defaults", zap.String("issuer", base))
	return fallback
}

func (i *Installer) enableOIDCInDb(dbPath string, discovery *oidcDiscovery, secret string) error {
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

	settings["authActiveAuthMethods"] = []string{"local", "openid"}
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
