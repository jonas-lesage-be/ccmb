package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"ccmb/internal/conv"
)

// Config holds the configuration settings for the application.
type Config struct {
	MaxFlattenerWorkerCount int
	MaxVideoWorkerCount     int

	EstFileCount        int
	FlatPathDelimiter   string
	EscapedDelimiter    string
	DecodePlaceholder   string
	VideoExtractTimeout time.Duration

	FilterExtensions map[string]bool
	ImageExtensions  map[string]bool
	VideoExtensions  map[string]bool
	VisualExtensions map[string]bool

	MaxZIPFileBytes  int64
	MaxPDFFileBytes  int64
	MaxTextFileBytes int64

	SourceDir            string
	TargetDir            string
	TargetDirPermissions os.FileMode
	TextFilePermissions  os.FileMode

	RunFlattener      bool
	RunFilter         bool
	RunVideoExtractor bool
	RunVisualMerger   bool
	RunTextMerger     bool
}

const (
	defaultMaxFlattenerWorkerCount = 8
	defaultMaxVideoWorkerCount     = 4

	defaultEstFileCount        = 128
	defaultFlatPathDelimiter   = "--"
	defaultDecodePlaceholder   = "\x00"
	defaultVideoExtractTimeout = 1 * time.Minute

	defaultFilterExtensions = ".ttf,.woff,.woff2"
	defaultImageExtensions  = ".jpg,.jpeg,.png,.webp,.tif,.tiff"
	defaultVideoExtensions  = "nil"
	defaultVisualExtensions = ".pdf," + defaultImageExtensions

	defaultMaxZIPFileBytes  int64 = 4 * conv.GiB
	defaultMaxPDFFileBytes  int64 = 128 * conv.MiB
	defaultMaxTextFileBytes int64 = 128 * conv.MiB

	defaultSourceDir                        = "."
	defaultTargetDir                        = "./_ccmb_output"
	defaultTargetDirPermissions os.FileMode = 0o700
	defaultTextFilePermissions  os.FileMode = 0o600

	defaultRunFlattener      = true
	defaultRunFilter         = true
	defaultRunVideoExtractor = true
	defaultRunVisualMerger   = true
	defaultRunTextMerger     = true
)

// Load reads the configuration from environment variables and returns a Config struct.
func Load() *Config {
	filePathDelimiter := EnvOr("FLAT_PATH_DELIMITER", defaultFlatPathDelimiter)

	return &Config{
		MaxFlattenerWorkerCount: EnvOr(
			"MAX_FLATTENER_WORKER_COUNT",
			defaultMaxFlattenerWorkerCount,
		),
		MaxVideoWorkerCount: EnvOr("MAX_VIDEO_WORKER_COUNT", defaultMaxVideoWorkerCount),

		EstFileCount:        EnvOr("EST_FILE_COUNT", defaultEstFileCount),
		FlatPathDelimiter:   filePathDelimiter,
		EscapedDelimiter:    EnvOr("ESCAPED_DELIMITER", filePathDelimiter+filePathDelimiter),
		DecodePlaceholder:   EnvOr("DECODE_PLACEHOLDER", defaultDecodePlaceholder),
		VideoExtractTimeout: EnvOr("VIDEO_EXTRACT_TIMEOUT", defaultVideoExtractTimeout),

		FilterExtensions: EnvMapOr("FILTER_EXTENSIONS", defaultFilterExtensions),
		ImageExtensions:  EnvMapOr("IMAGE_EXTENSIONS", defaultImageExtensions),
		VideoExtensions:  EnvMapOr("VIDEO_EXTENSIONS", defaultVideoExtensions),
		VisualExtensions: EnvMapOr("VISUAL_EXTENSIONS", defaultVisualExtensions),

		MaxZIPFileBytes:  EnvOr("MAX_ZIP_FILE_BYTES", defaultMaxZIPFileBytes),
		MaxPDFFileBytes:  EnvOr("MAX_PDF_FILE_BYTES", defaultMaxPDFFileBytes),
		MaxTextFileBytes: EnvOr("MAX_TEXT_FILE_BYTES", defaultMaxTextFileBytes),

		SourceDir:            EnvOr("SOURCE_DIR", defaultSourceDir),
		TargetDir:            EnvOr("TARGET_DIR", defaultTargetDir),
		TargetDirPermissions: EnvOr("TARGET_DIR_PERMISSIONS", defaultTargetDirPermissions),
		TextFilePermissions:  EnvOr("TEXT_FILE_PERMISSIONS", defaultTextFilePermissions),

		RunFlattener:      EnvOr("RUN_FLATTENER", defaultRunFlattener),
		RunFilter:         EnvOr("RUN_FILTER", defaultRunFilter),
		RunVideoExtractor: EnvOr("RUN_VIDEO_EXTRACTOR", defaultRunVideoExtractor),
		RunVisualMerger:   EnvOr("RUN_VISUAL_MERGER", defaultRunVisualMerger),
		RunTextMerger:     EnvOr("RUN_TEXT_MERGER", defaultRunTextMerger),
	}
}

// ShouldIgnore determines if a directory component contains macOS junk files.
func ShouldIgnore(pathStr string) bool {
	parts := strings.SplitSeq(filepath.ToSlash(pathStr), "/")

	for part := range parts {
		if part == "__MACOSX" || strings.HasPrefix(part, ".DS_") {
			return true
		}
	}

	return false
}

// EncodeFlatName converts a raw path into a flattened name using the configured delimiters.
func EncodeFlatName(rawPath, flatPathDelimiter, escapedDelimiter string) string {
	cleanPath := filepath.ToSlash(rawPath)
	escaped := strings.ReplaceAll(cleanPath, flatPathDelimiter, escapedDelimiter)
	return strings.ReplaceAll(escaped, "/", flatPathDelimiter)
}

// DecodeFlatName converts flattened names back into paths with forward slashes ("/").
func DecodeFlatName(
	flatName, escapedDelimiter, decodePlaceholder, flatPathDelimiter string,
) string {
	t := strings.ReplaceAll(flatName, escapedDelimiter, decodePlaceholder)
	t = strings.ReplaceAll(t, flatPathDelimiter, "/")
	return strings.ReplaceAll(t, decodePlaceholder, flatPathDelimiter)
}
