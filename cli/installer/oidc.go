package installer

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"
	"io"
	_ "modernc.org/sqlite"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	routerBasePath         = "/" + App
	oidcWebCallbackPath    = routerBasePath + "/auth/openid/callback"
	oidcMobileRedirectPath = routerBasePath + "/auth/openid/mobile-redirect"
	serverSettingsKey      = "server-settings"
	adminUsername          = "admin"
	defaultLibraryName     = "Books"
	defaultLibraryDir      = "library"
)

type absStatus struct {
	IsInit bool `json:"isInit"`
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

func (o *Oidc) Initialize(storageDir string) error {
	isInit, err := o.waitForStatus()
	if err != nil {
		return err
	}
	if isInit {
		return nil
	}
	password, err := o.createRootUser()
	if err != nil {
		return err
	}
	token, err := o.login(adminUsername, password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if err := o.createLibrary(token, defaultLibraryName, path.Join(storageDir, defaultLibraryDir)); err != nil {
		return fmt.Errorf("create default library: %w", err)
	}
	return nil
}

func (o *Oidc) ConfigureApp() error {
	dbPath := path.Join(o.dataDir, "config", "absdatabase.sqlite")
	secret, err := o.platformClient.RegisterOIDCClient(App, []string{oidcWebCallbackPath, oidcMobileRedirectPath}, true, "client_secret_basic")
	if err != nil {
		return fmt.Errorf("oidc register: %w", err)
	}
	authUrl, err := o.platformClient.GetAppUrl("auth")
	if err != nil {
		return err
	}
	if err := o.enableOIDCInDb(dbPath, authUrl, secret); err != nil {
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

func (o *Oidc) createRootUser() (string, error) {
	password, err := randomPassword()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]interface{}{
		"newRoot": map[string]string{"username": adminUsername, "password": password},
	})
	if err != nil {
		return "", err
	}
	resp, err := o.client.Post("http://localhost/init", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("init returned %s", resp.Status)
	}
	passwordFile := path.Join(o.dataDir, "initial_admin_password")
	if err := os.WriteFile(passwordFile, []byte(password+"\n"), 0600); err != nil {
		return "", err
	}
	o.logger.Info("created root admin user", zap.String("username", adminUsername), zap.String("password_file", passwordFile))
	return password, nil
}

func (o *Oidc) login(username, password string) (string, error) {
	payload, err := json.Marshal(map[string]string{"username": username, "password": password})
	if err != nil {
		return "", err
	}
	resp, err := o.client.Post("http://localhost/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned %s", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	var lr struct {
		User struct {
			AccessToken string `json:"accessToken"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &lr); err != nil {
		return "", err
	}
	if lr.User.AccessToken == "" {
		return "", fmt.Errorf("login returned no access token")
	}
	return lr.User.AccessToken, nil
}

func (o *Oidc) createLibrary(token, name, folderPath string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"name":      name,
		"mediaType": "book",
		"folders":   []map[string]string{{"fullPath": folderPath}},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "http://localhost/api/libraries", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create library returned %s: %s", resp.Status, string(body))
	}
	o.logger.Info("created default library", zap.String("name", name), zap.String("path", folderPath))
	return nil
}

func randomPassword() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (o *Oidc) enableOIDCInDb(dbPath string, authUrl string, secret string) error {
	base := strings.TrimRight(authUrl, "/")
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
	settings["authOpenIDIssuerURL"] = base
	settings["authOpenIDAuthorizationURL"] = base + "/api/oidc/authorization"
	settings["authOpenIDTokenURL"] = base + "/api/oidc/token"
	settings["authOpenIDUserInfoURL"] = base + "/api/oidc/userinfo"
	settings["authOpenIDJwksURL"] = base + "/jwks.json"
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
