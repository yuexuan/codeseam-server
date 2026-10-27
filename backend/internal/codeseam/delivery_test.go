package codeseam

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeliveryPublicFilesAndBoundaries(t *testing.T) {
	root := t.TempDir()
	portal := filepath.Join(root, "portal")
	if err := os.Mkdir(portal, 0700); err != nil {
		t.Fatal(err)
	}
	for p, v := range map[string]string{filepath.Join(portal, "index.html"): "portal-production", filepath.Join(root, "private.txt"): "must-not-leak"} {
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
		status       int
		body         string
	}{
		{"GET", "/portal/", 200, "portal-production"},
		{"GET", "/portal", 307, ""},
		{"GET", "/portal/../private.txt", 404, ""},
		{"GET", "/portal/%2e%2e/private.txt", 404, ""},
		{"GET", "/portal/missing.js", 404, ""},
		{"GET", "/downloads/codeseam-windows.exe", 404, ""},
		{"POST", "/portal/", 405, ""},
		{"GET", "/api/v1/auth/me", 418, "operations"},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(test.method, test.path, nil))
		if w.Code != test.status {
			t.Errorf("%s %s: HTTP %d, want %d", test.method, test.path, w.Code, test.status)
		}
		if test.body != "" && w.Body.String() != test.body {
			t.Errorf("%s: unexpected response body", test.path)
		}
	}
}
