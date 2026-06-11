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
	"path/filepath"
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

type Abs struct {
	platformClient *platform.Client
	logger         *zap.Logger
	dataDir        string
	client         *http.Client
}

func NewAbs(platformClient *platform.Client, logger *zap.Logger, dataDir string) *Abs {
	socket := path.Join(dataDir, "audiobookshelf.sock")
	return &Abs{
		platformClient: platformClient,
		logger:         logger,
		dataDir:        dataDir,
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", socket)
				},
			},
		},
	}
}

func (a *Abs) Initialize(storageDir string) error {
	if err := a.waitForReady(); err != nil {
		return err
	}
	password, err := a.createRootUser()
	if err != nil {
		return err
	}
	token, err := a.login(adminUsername, password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	libraryPath, err := filepath.EvalSymlinks(path.Join(storageDir, defaultLibraryDir))
	if err != nil {
		return fmt.Errorf("resolve library path: %w", err)
	}
	if err := a.createLibrary(token, defaultLibraryName, libraryPath); err != nil {
		return fmt.Errorf("create default library: %w", err)
	}
	return nil
}

func (a *Abs) ConfigureApp() error {
	dbPath := path.Join(a.dataDir, "config", "absdatabase.sqlite")
	secret, err := a.platformClient.RegisterOIDCClient(App, []string{oidcWebCallbackPath, oidcMobileRedirectPath}, true, "client_secret_basic")
	if err != nil {
		return fmt.Errorf("oidc register: %w", err)
	}
	authUrl, err := a.platformClient.GetAppUrl("auth")
	if err != nil {
		return err
	}
	if err := a.enableOIDCInDb(dbPath, authUrl, secret); err != nil {
		return fmt.Errorf("oidc update db: %w", err)
	}
	return a.platformClient.RestartService(fmt.Sprint(App, ".abs"))
}

func (a *Abs) waitForReady() error {
	for attempt := 0; attempt < 60; attempt++ {
		resp, err := a.client.Get("http://localhost/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("audiobookshelf did not become ready")
}

func (a *Abs) createRootUser() (string, error) {
	password, err := randomPassword()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(newRootRequest{NewRoot: credentials{Username: adminUsername, Password: password}})
	if err != nil {
		return "", err
	}
	resp, err := a.client.Post("http://localhost/init", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("init returned %s", resp.Status)
	}
	passwordFile := path.Join(a.dataDir, "initial_admin_password")
	if err := os.WriteFile(passwordFile, []byte(password+"\n"), 0600); err != nil {
		return "", err
	}
	a.logger.Info("created root admin user", zap.String("username", adminUsername), zap.String("password_file", passwordFile))
	return password, nil
}

func (a *Abs) login(username, password string) (string, error) {
	payload, err := json.Marshal(credentials{Username: username, Password: password})
	if err != nil {
		return "", err
	}
	resp, err := a.client.Post("http://localhost/login", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned %s", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	var lr loginResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return "", err
	}
	if lr.User.AccessToken == "" {
		return "", fmt.Errorf("login returned no access token")
	}
	return lr.User.AccessToken, nil
}

func (a *Abs) createLibrary(token, name, folderPath string) error {
	payload, err := json.Marshal(createLibraryRequest{
		Name:      name,
		MediaType: "book",
		Folders:   []libraryFolder{{FullPath: folderPath}},
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
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create library returned %s: %s", resp.Status, string(body))
	}
	a.logger.Info("created default library", zap.String("name", name), zap.String("path", folderPath))
	return nil
}

func randomPassword() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (a *Abs) enableOIDCInDb(dbPath string, authUrl string, secret string) error {
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
