package api

import (
	"net/http"
	"os"
	"path/filepath"
)

// CAFilePath is the on-disk path the proxy CA cert is mounted at.
// Default mirrors the Node.js layout (/work/cassl/rootCA.crt).
var CAFilePath = "/work/cassl/rootCA.crt"

// DownloadCAHandler is GET /api/v1/download_ca
func DownloadCAHandler(w http.ResponseWriter, _ *http.Request) {
	path := os.Getenv("CA_FILE_PATH")
	if path == "" {
		path = CAFilePath
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is configured by operator
	if err != nil {
		JSONErr(w, http.StatusNotFound, "CA file not found")
		return
	}
	w.Header().Set("Content-Type", "application/x-x509-ca-cert")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(path)+`"`)
	_, _ = w.Write(data)
}
