package document

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var errLibreOfficeNotFound = errors.New(
	"libreOffice binary (soffice) not found in PATH or default installation locations",
)

func (c *Converter) convertToIntermediate(ctx context.Context, path, ext string) (string, error) {
	loBinary, err := findLibreOfficeBinary()
	if err != nil {
		return "", err
	}

	userProfileDir, err := os.MkdirTemp("", "ccmb_lo_profile_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary profile for LibreOffice: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(userProfileDir); err != nil {
			slog.Error("failed to remove temporary profile directory", "err", err)
		}
	}()

	profileURL := "file:///" + filepath.ToSlash(userProfileDir)
	targetFormat := loConversionTargetFormat(ext)
	targetExt := "." + targetFormat

	args := []string{
		"-env:UserInstallation=" + profileURL,
		"--headless",
		"--convert-to", targetFormat,
		"--outdir", c.Dir,
		path,
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, loBinary, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("libreoffice conversion command failed: %w", err)
	}

	filename := filepath.Base(path)
	standardPath := filepath.Join(c.Dir, strings.TrimSuffix(filename, ext)+targetExt)

	generatedName := filename + c.FlatPathDelimiter + targetExt
	generatedPath := filepath.Join(c.Dir, generatedName)

	if err := os.Rename(standardPath, generatedPath); err != nil {
		return "", fmt.Errorf("failed to apply delimiter extension to generated PDF: %w", err)
	}

	return generatedPath, nil
}

func isLOConvertible(ext string) bool {
	switch ext {
	case ".doc", ".docm", ".xls", ".xlsx", ".xlsm", ".ods", ".ppt", ".pptx", ".pptm", ".odp":
		return true
	default:
		return false
	}
}

func findLibreOfficeBinary() (string, error) {
	for _, name := range loSearchNames() {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	switch runtime.GOOS {
	case "windows":
		defaultPath := `C:\Program Files\LibreOffice\program\soffice.exe`
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath, nil
		}

	case "darwin":
		defaultPath := "/Applications/LibreOffice.app/Contents/MacOS/soffice"
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath, nil
		}

	case "linux":
		commonPaths := []string{
			"/usr/bin/soffice",
			"/var/lib/flatpak/exports/bin/org.libreoffice.LibreOffice",
			"/opt/libreoffice/program/soffice",
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", errLibreOfficeNotFound
}

func loSearchNames() []string {
	if runtime.GOOS == "linux" {
		return []string{"soffice", "libreoffice.soffice"}
	}
	return []string{"soffice"}
}

func loConversionTargetFormat(ext string) string {
	switch ext {
	case ".doc", ".docm":
		return "docx"
	case ".xls", ".xlsx", ".xlsm", ".ods":
		return "pdf"
	case ".ppt", ".pptx", ".pptm", ".odp":
		return "pdf"
	default:
		return "nil"
	}
}
