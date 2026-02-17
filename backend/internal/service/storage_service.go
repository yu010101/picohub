package service

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrFileTooLarge    = errors.New("file exceeds maximum size")
	ErrInvalidPackage  = errors.New("invalid skill package")
	ErrNoManifest      = errors.New("manifest.json not found in package")
	ErrSymlinkDetected = errors.New("symbolic links are not allowed")
)

type StorageService struct {
	uploadDir   string
	maxFileSize int64
}

func NewStorageService(uploadDir string, maxFileSize int64) *StorageService {
	os.MkdirAll(uploadDir, 0o755)
	return &StorageService{uploadDir: uploadDir, maxFileSize: maxFileSize}
}

type SkillManifest struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Author      string   `json:"author"`
	EntryPoint  string   `json:"entry_point"`
}

func (s *StorageService) SaveFile(src io.Reader, originalName string) (filePath string, fileHash string, manifest *SkillManifest, err error) {
	// Save to temp file first
	tmpFile, err := os.CreateTemp(s.uploadDir, "upload-*.zip")
	if err != nil {
		return "", "", nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Copy with size limit
	hasher := sha256.New()
	limited := io.LimitReader(src, s.maxFileSize+1)
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), limited)
	tmpFile.Close()
	if err != nil {
		return "", "", nil, fmt.Errorf("save file: %w", err)
	}
	if written > s.maxFileSize {
		return "", "", nil, ErrFileTooLarge
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	// Validate zip and extract manifest
	manifest, err = s.validatePackage(tmpPath)
	if err != nil {
		return "", "", nil, err
	}

	// Rename to UUID-based name
	finalName := uuid.New().String() + ".zip"
	finalPath := filepath.Join(s.uploadDir, finalName)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", "", nil, fmt.Errorf("rename file: %w", err)
	}

	return finalPath, hash, manifest, nil
}

func (s *StorageService) validatePackage(path string) (*SkillManifest, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, ErrInvalidPackage
	}
	defer r.Close()

	var manifestFile *zip.File
	for _, f := range r.File {
		// Check for symlinks
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			return nil, ErrSymlinkDetected
		}
		// Check for path traversal
		if strings.Contains(f.Name, "..") {
			return nil, ErrInvalidPackage
		}
		// Find manifest.json (at root or one level deep)
		base := filepath.Base(f.Name)
		depth := len(strings.Split(strings.TrimSuffix(f.Name, "/"), "/"))
		if base == "manifest.json" && depth <= 2 {
			manifestFile = f
		}
	}

	if manifestFile == nil {
		return nil, ErrNoManifest
	}

	rc, err := manifestFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open manifest: %w", err)
	}
	defer rc.Close()

	var manifest SkillManifest
	if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if manifest.Name == "" || manifest.Slug == "" {
		return nil, fmt.Errorf("manifest must contain name and slug")
	}

	return &manifest, nil
}

func (s *StorageService) GetFilePath(storedPath string) string {
	return storedPath
}

func (s *StorageService) Delete(filePath string) error {
	if filePath == "" {
		return nil
	}
	return os.Remove(filePath)
}
