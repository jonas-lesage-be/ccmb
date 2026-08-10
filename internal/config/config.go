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
	SourceDir            string
	TargetDir            string
	TargetDirPermissions os.FileMode
	TextFilePermissions  os.FileMode

	MaxFlattenerWorkers    int
	MaxImageWorkers        int
	MaxVideoWorkers        int
	ImageConversionTimeout time.Duration
	VideoExtractTimeout    time.Duration

	EstFileCount      int
	FlatPathDelimiter string
	EscapedDelimiter  string
	DecodePlaceholder string

	MaxArchiveFileBytes int64
	MaxPDFFileBytes     int64
	MaxTextFileBytes    int64

	FilterExtensions         map[string]bool
	ImageExtensions          map[string]bool
	SupportedImageExtensions map[string]bool
	VideoExtensions          map[string]bool
	VisualExtensions         map[string]bool

	RunFlattener      bool
	RunTARFlattener   bool
	RunFilter         bool
	RunImageConverter bool
	RunVideoExtractor bool
	RunVisualMerger   bool
	RunTextMerger     bool
}

const (
	defaultSourceDir                        = "."
	defaultTargetDir                        = "./_ccmb_output"
	defaultTargetDirPermissions os.FileMode = 0o700
	defaultTextFilePermissions  os.FileMode = 0o600

	defaultMaxFlattenerWorkers    = 8
	defaultMaxImageWorkers        = 8
	defaultMaxVideoWorkers        = 4
	defaultImageConversionTimeout = 15 * time.Second
	defaultVideoExtractTimeout    = 1 * time.Minute

	defaultEstFileCount      = 128
	defaultFlatPathDelimiter = "--"
	defaultDecodePlaceholder = "\x00"

	defaultMaxArchiveFileBytes int64 = 4 * conv.GiB
	defaultMaxPDFFileBytes     int64 = 48 * conv.MiB
	defaultMaxTextFileBytes    int64 = 2 * conv.MiB

	defaultImageExtensions          = "nil"
	defaultSupportedImageExtensions = ".jpg,.jpeg,.png,.webp,.tif,.tiff"
	defaultVideoExtensions          = "nil"
	defaultVisualExtensions         = ".pdf," + defaultSupportedImageExtensions

	defaultRunFlattener      = true
	defaultRunTARFlattener   = !isWindows
	defaultRunFilter         = true
	defaultRunImageConverter = true
	defaultRunVideoExtractor = true
	defaultRunVisualMerger   = true
	defaultRunTextMerger     = true
)

// Load reads the configuration from environment variables and returns a Config struct.
func Load() *Config {
	filePathDelimiter := EnvOr("FLAT_PATH_DELIMITER", defaultFlatPathDelimiter)

	return &Config{
		SourceDir:            EnvOr("SOURCE_DIR", defaultSourceDir),
		TargetDir:            EnvOr("TARGET_DIR", defaultTargetDir),
		TargetDirPermissions: EnvOr("TARGET_DIR_PERMISSIONS", defaultTargetDirPermissions),
		TextFilePermissions:  EnvOr("TEXT_FILE_PERMISSIONS", defaultTextFilePermissions),

		MaxFlattenerWorkers: EnvOr(
			"MAX_FLATTENER_WORKERS",
			defaultMaxFlattenerWorkers,
		),
		MaxImageWorkers:        EnvOr("MAX_IMAGE_WORKERS", defaultMaxImageWorkers),
		MaxVideoWorkers:        EnvOr("MAX_VIDEO_WORKERS", defaultMaxVideoWorkers),
		ImageConversionTimeout: EnvOr("IMAGE_CONVERSION_TIMEOUT", defaultImageConversionTimeout),
		VideoExtractTimeout:    EnvOr("VIDEO_EXTRACT_TIMEOUT", defaultVideoExtractTimeout),

		EstFileCount:      EnvOr("EST_FILE_COUNT", defaultEstFileCount),
		FlatPathDelimiter: filePathDelimiter,
		EscapedDelimiter:  EnvOr("ESCAPED_DELIMITER", filePathDelimiter+filePathDelimiter),
		DecodePlaceholder: EnvOr("DECODE_PLACEHOLDER", defaultDecodePlaceholder),

		MaxArchiveFileBytes: EnvOr("MAX_ARCHIVE_FILE_BYTES", defaultMaxArchiveFileBytes),
		MaxPDFFileBytes:     EnvOr("MAX_PDF_FILE_BYTES", defaultMaxPDFFileBytes),
		MaxTextFileBytes:    EnvOr("MAX_TEXT_FILE_BYTES", defaultMaxTextFileBytes),

		FilterExtensions: EnvMapOr("FILTER_EXTENSIONS", defaultFilterExtensions),
		ImageExtensions:  EnvMapOr("IMAGE_EXTENSIONS", defaultImageExtensions),
		SupportedImageExtensions: EnvMapOr(
			"SUPPORTED_IMAGE_EXTENSIONS",
			defaultSupportedImageExtensions,
		),
		VideoExtensions:  EnvMapOr("VIDEO_EXTENSIONS", defaultVideoExtensions),
		VisualExtensions: EnvMapOr("VISUAL_EXTENSIONS", defaultVisualExtensions),

		RunFlattener:      EnvOr("RUN_FLATTENER", defaultRunFlattener),
		RunTARFlattener:   EnvOr("RUN_TAR_FLATTENER", defaultRunTARFlattener),
		RunFilter:         EnvOr("RUN_FILTER", defaultRunFilter),
		RunImageConverter: EnvOr("RUN_IMAGE_CONVERTER", defaultRunImageConverter),
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
