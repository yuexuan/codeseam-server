package codeseam

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeliveryPathRoutes(t *testing.T) {
	root := t.TempDir()
	portal := filepath.Join(root, "portal")
	if err := os.MkdirAll(filepath.Join(portal, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for p, v := range map[string]string{
		filepath.Join(portal, "index.html"): "codeseam-production",
		filepath.Join(portal, "assets", "index-abc.js"): "codeseam-bundle",
		filepath.Join(root, "private.txt"): "must-not-leak",
	} {
		if err := os.WriteFile(p, []byte(v), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CODESEAM_PORTAL_DIR", portal)
	t.Setenv("CODESEAM_WINDOWS_INSTALLER", "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Delivery())
	r.NoRoute(func(c *gin.Context) { c.String(418, "operations") })
	for _, test := range []struct {
		method, path string
		status int
		body string
	}{
		{"GET", "/", 200, "codeseam-production"},
		{"GET", "/login", 200, "codeseam-production"},
		{"GET", "/register", 200, "codeseam-production"},
		{"GET", "/forgot", 200, "codeseam-production"},
		{"GET", "/app", 200, "codeseam-production"},
		{"GET", "/app/usage", 200, "codeseam-production"},
		{"GET", "/app/service", 200, "codeseam-production"},
		{"GET", "/assets/index-abc.js", 200, "codeseam-bundle"},
		{"GET", "/assets/missing.js", 404, ""},
		{"GET", "/assets/%2e%2e/private.txt", 404, ""},
		{"GET", "/portal/", 418, "operations"},
		{"GET", "/downloads/codeseam-windows.exe", 404, ""},
		{"POST", "/app", 405, ""},
		{"GET", "/api/v1/auth/me", 418, "operations"},
	} {
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))
		if recorder.Code != test.status {
			t.Errorf("%s %s: HTTP %d, want %d", test.method, test.path, recorder.Code, test.status)
		}
		if test.body != "" && recorder.Body.String() != test.body {
			t.Errorf("%s: unexpected response body %q", test.path, recorder.Body.String())
		}
	}
}
