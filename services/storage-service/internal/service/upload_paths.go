package service

import (
	"path/filepath"
	"strings"
)

// normalizeUploadSubdir converts API-style upload paths (e.g. "/uploads/profile")
// into a subdirectory relative to the local upload base (e.g. "profile").
// An empty return value means the default mime/date layout under uploads/.
func normalizeUploadSubdir(uploadPath string) string {
	p := strings.TrimSpace(uploadPath)
	p = strings.Trim(p, "/")
	if strings.HasPrefix(p, "uploads/") {
		p = strings.TrimPrefix(p, "uploads/")
	} else if p == "uploads" {
		p = ""
	}
	return strings.Trim(p, "/")
}

// resolveChunkLocalPath maps an assembled relative path to a writable filesystem path.
func resolveChunkLocalPath(uploadBaseDir, relativePath string, customUpload bool) string {
	if customUpload {
		return filepath.Join(uploadBaseDir, relativePath)
	}
	return relativePath
}

// resolveChunkPublicDir returns the directory path exposed to API clients.
func resolveChunkPublicDir(relativePath, uploadSubdir string, customUpload bool) string {
	if customUpload {
		dir := "/uploads/" + strings.ReplaceAll(uploadSubdir, "\\", "/")
		if !strings.HasSuffix(dir, "/") {
			dir += "/"
		}
		return dir
	}

	pathDir := filepath.Dir(relativePath)
	pathDir = strings.ReplaceAll(pathDir, "\\", "/")
	if !strings.HasSuffix(pathDir, "/") {
		pathDir += "/"
	}
	return pathDir
}
