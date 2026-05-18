package service

import (
	"path/filepath"
	"testing"
)

func TestNormalizeUploadSubdir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "profile API path", input: "/uploads/profile", want: "profile"},
		{name: "profile without leading slash", input: "uploads/profile", want: "profile"},
		{name: "kyc path", input: "/uploads/kyc", want: "kyc"},
		{name: "subdir only", input: "profile", want: "profile"},
		{name: "empty", input: "", want: ""},
		{name: "uploads root", input: "/uploads", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeUploadSubdir(tt.input); got != tt.want {
				t.Fatalf("normalizeUploadSubdir(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveChunkLocalPath(t *testing.T) {
	base := filepath.Join("data", "uploads")
	got := resolveChunkLocalPath(base, filepath.Join("profile", "abc.jpg"), true)
	want := filepath.Join(base, "profile", "abc.jpg")
	if got != want {
		t.Fatalf("custom upload local path = %q, want %q", got, want)
	}

	defaultPath := filepath.Join("uploads", "image-jpeg", "2024-01-01", "abc.jpg")
	if got := resolveChunkLocalPath(base, defaultPath, false); got != defaultPath {
		t.Fatalf("default upload local path = %q, want %q", got, defaultPath)
	}
}

func TestResolveChunkPublicDir(t *testing.T) {
	if got := resolveChunkPublicDir("", "profile", true); got != "/uploads/profile/" {
		t.Fatalf("custom public dir = %q, want /uploads/profile/", got)
	}

	relative := filepath.Join("uploads", "image-jpeg", "2024-01-01", "abc.jpg")
	if got := resolveChunkPublicDir(relative, "", false); got != "uploads/image-jpeg/2024-01-01/" {
		t.Fatalf("default public dir = %q, want uploads/image-jpeg/2024-01-01/", got)
	}
}
