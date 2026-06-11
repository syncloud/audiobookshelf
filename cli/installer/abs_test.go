package installer

import (
	"database/sql"
	"encoding/json"
	"path"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEnableOIDCInDb(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "absdatabase.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE settings (key TEXT PRIMARY KEY, value JSON)`); err != nil {
		t.Fatal(err)
	}
	seed := `{"id":"server-settings","authActiveAuthMethods":["local"],"authOpenIDIssuerURL":null}`
	if _, err := db.Exec(`INSERT INTO settings(key,value) VALUES(?,?)`, serverSettingsKey, seed); err != nil {
		t.Fatal(err)
	}
	db.Close()

	o := &Abs{}
	if err := o.enableOIDCInDb(dbPath, "https://auth.example.com/", "s3cr3t"); err != nil {
		t.Fatal(err)
	}

	db, _ = sql.Open("sqlite", dbPath)
	defer db.Close()
	var raw string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key=?`, serverSettingsKey).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var s map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("value is not valid json: %v", err)
	}
	methods, _ := json.Marshal(s["authActiveAuthMethods"])
	if string(methods) != `["openid"]` {
		t.Fatalf("auth methods = %s", methods)
	}
	if s["authOpenIDClientSecret"] != "s3cr3t" || s["authOpenIDClientID"] != App {
		t.Fatalf("client creds not set: %v / %v", s["authOpenIDClientID"], s["authOpenIDClientSecret"])
	}
	if s["authOpenIDIssuerURL"] != "https://auth.example.com" || s["authOpenIDTokenURL"] != "https://auth.example.com/api/oidc/token" || s["authOpenIDJwksURL"] != "https://auth.example.com/jwks.json" || s["id"] != "server-settings" {
		t.Fatalf("endpoints/id wrong: %v", s)
	}
	if _, ok := s["authOpenIDLogoutURL"]; ok {
		t.Fatalf("logout url should not be set (authelia has no end_session_endpoint)")
	}
	if s["authOpenIDAutoRegister"] != true {
		t.Fatalf("autoRegister not set")
	}
	if s["authOpenIDGroupClaim"] != "groups" || s["authOpenIDAdminGroups"] != "syncloud" || s["authOpenIDGroupDefaultRole"] != "user" {
		t.Fatalf("group mapping not set: claim=%v admin=%v default=%v", s["authOpenIDGroupClaim"], s["authOpenIDAdminGroups"], s["authOpenIDGroupDefaultRole"])
	}
	if s["authOpenIDSubfolderForRedirectURLs"] != routerBasePath {
		t.Fatalf("redirect subfolder = %v, want %v", s["authOpenIDSubfolderForRedirectURLs"], routerBasePath)
	}
}
