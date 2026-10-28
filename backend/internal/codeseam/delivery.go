package codeseam

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Delivery serves CodeSeam's public site at /, its user application at /app,
// and static Vite assets. The existing operations API and admin routes remain separate.
func Delivery() gin.HandlerFunc {
	portal := os.Getenv("CODESEAM_PORTAL_DIR")
	installer := os.Getenv("CODESEAM_WINDOWS_INSTALLER")
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if p == "/downloads/codeseam-windows.exe" {
			if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				c.AbortWithStatus(http.StatusMethodNotAllowed)
				return
			}
			if info, err := os.Stat(installer); err != nil || !info.Mode().IsRegular() {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Header("Content-Disposition", "attachment; filename=CodeSeam-windows-x64-setup.exe")
			c.Header("Cache-Control", "no-cache")
			c.File(installer)
			c.Abort()
			return
		}
		if portal == "" {
			c.Next()
			return
		}
		var rel string
		switch {
		case p == "/" || p == "/index.html" || p == "/login" || p == "/register" || p == "/forgot" ||
			p == "/app" || strings.HasPrefix(p, "/app/"):
			rel = "index.html"
		case strings.HasPrefix(p, "/assets/"):
			rel = strings.TrimPrefix(p, "/")
		default:
			c.Next()
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		if !filepath.IsLocal(filepath.FromSlash(rel)) || strings.Contains(rel, "\\") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		target := filepath.Join(portal, filepath.FromSlash(rel))
		if info, err := os.Stat(target); err != nil || !info.Mode().IsRegular() {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if rel == "index.html" {
			c.Header("Cache-Control", "no-cache")
		} else {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.File(target)
		c.Abort()
	}
}
