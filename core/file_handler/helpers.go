package file_handler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CheckIPEFile(path string) bool {
	return filepath.Ext(path) == ".ipe"
}

func CheckValidIPEDirectory(path string) (*Manifest, error) {
	manifestPath := filepath.Join(path, "manifest.json")
	manifest, err := ParseManifest(manifestPath)
	return manifest, err
}

func ExtractIPE(ipePath string, destination string) error {
	archive, err := zip.OpenReader(ipePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		if f.Name == "" {
			continue
		}

		// Clean relative path
		destPath := filepath.Join(destination, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("zip file contains invalid path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			log.Printf("creating directory: %s\n", destPath)
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Ensure parent folders exist
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", destPath, err)
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer dstFile.Close()

		srcFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file inside zip: %w", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to write file: %w", err)
		}
		srcFile.Close()
	}

	return nil
}

func WriteManifest(manifestPath string, manifest *Manifest) error {
	// Marshal the Manifest into pretty-printed JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to the provided path
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}