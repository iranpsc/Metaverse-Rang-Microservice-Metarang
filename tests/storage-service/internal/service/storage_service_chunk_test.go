package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"metarang/storage-service/internal/ftp"
	"metarang/storage-service/internal/service"
)

func TestHandleChunkUploadProfilePath(t *testing.T) {
	tempDir := t.TempDir()
	uploadBase := filepath.Join(tempDir, "uploads")
	chunkTemp := filepath.Join(tempDir, "chunks")

	chunkManager, err := service.NewChunkManager(chunkTemp)
	if err != nil {
		t.Fatalf("NewChunkManager: %v", err)
	}

	mockFTP := ftp.NewMockFTPClient(filepath.Join(tempDir, "ftp"), "http://localhost/uploads")
	svc := service.NewStorageService(mockFTP, chunkManager, uploadBase)

	finished, progress, publicDir, filename, mimeType, err := svc.HandleChunkUpload(
		"test-upload-1",
		"photo.jpg",
		"image/jpeg",
		[]byte("fake-jpeg-data"),
		0,
		1,
		int64(len("fake-jpeg-data")),
		"/uploads/profile",
	)
	if err != nil {
		t.Fatalf("HandleChunkUpload: %v", err)
	}
	if !finished || progress != 100.0 {
		t.Fatalf("unexpected progress: finished=%v progress=%v", finished, progress)
	}
	if publicDir != "/uploads/profile/" {
		t.Fatalf("publicDir = %q, want /uploads/profile/", publicDir)
	}
	if filename == "" {
		t.Fatal("expected generated filename")
	}
	if mimeType != "image/jpeg" {
		t.Fatalf("mimeType = %q, want image/jpeg", mimeType)
	}

	localFile := filepath.Join(uploadBase, "profile", filename)
	if _, err := os.Stat(localFile); err != nil {
		t.Fatalf("expected file at %s: %v", localFile, err)
	}
}
