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
	SourceDir            string      `mapstructure:"source_dir"`
	TargetDir            string      `mapstructure:"target_dir"`
	TargetDirPermissions os.FileMode `mapstructure:"target_dir_permissions"`
	TextFilePermissions  os.FileMode `mapstructure:"text_file_permissions"`

	MaxFlattenerWorkers    int           `mapstructure:"max_flattener_workers"`
	MaxDocumentWorkers     int           `mapstructure:"max_document_workers"`
	MaxImageWorkers        int           `mapstructure:"max_image_workers"`
	MaxVideoWorkers        int           `mapstructure:"max_video_workers"`
	ImageConversionTimeout time.Duration `mapstructure:"image_conversion_timeout"`
	VideoExtractTimeout    time.Duration `mapstructure:"video_extract_timeout"`
	SvgCanvasResolution    float64       `mapstructure:"svg_canvas_resolution"`

	EstFileCount      int    `mapstructure:"est_file_count"`
	FlatPathDelimiter string `mapstructure:"flat_path_delimiter"`
	EscapedDelimiter  string `mapstructure:"escaped_delimiter"`
	DecodePlaceholder string `mapstructure:"decode_placeholder"`

	MaxArchiveFileBytes int64 `mapstructure:"max_archive_file_bytes"`
	MaxPDFFileBytes     int64 `mapstructure:"max_pdf_file_bytes"`
	MaxTextFileBytes    int64 `mapstructure:"max_text_file_bytes"`

	FilterExtensions         map[string]bool `mapstructure:"filter_extensions"`
	DocumentExtensions       map[string]bool `mapstructure:"document_extensions"`
	ImageExtensions          map[string]bool `mapstructure:"image_extensions"`
	SupportedImageExtensions map[string]bool `mapstructure:"supported_image_extensions"`
	VideoExtensions          map[string]bool `mapstructure:"video_extensions"`
	VisualExtensions         map[string]bool `mapstructure:"visual_extensions"`

	Verbose                  bool `mapstructure:"verbose"`
	SkipFlattener            bool `mapstructure:"skip_flattener"`
	SkipTARFlattener         bool `mapstructure:"skip_tar_flattener"`
	SkipTARFlattenerExplicit bool `mapstructure:"-"`
	SkipFilter               bool `mapstructure:"skip_filter"`
	SkipDocumentConverter    bool `mapstructure:"skip_document_converter"`
	SkipImageConverter       bool `mapstructure:"skip_image_converter"`
	SkipVideoExtractor       bool `mapstructure:"skip_video_extractor"`
	SkipVisualMerger         bool `mapstructure:"skip_visual_merger"`
	SkipTextMerger           bool `mapstructure:"skip_text_merger"`
}

const (
	defaultSourceDir                        = "."
	defaultTargetDir                        = "./_ccmb_output"
	defaultTargetDirPermissions os.FileMode = 0o700
	defaultTextFilePermissions  os.FileMode = 0o600

	defaultMaxFlattenerWorkers    = 8
	defaultMaxDocumentWorkers     = 8
	defaultMaxImageWorkers        = 8
	defaultMaxVideoWorkers        = 4
	defaultImageConversionTimeout = 15 * time.Second
	defaultVideoExtractTimeout    = 1 * time.Minute
	defaultSvgCanvasResolution    = 300 / 25.4

	defaultEstFileCount      = 128
	defaultFlatPathDelimiter = "--"
	defaultDecodePlaceholder = "\x00"

	defaultMaxArchiveFileBytes int64 = 4 * conv.GiB
	defaultMaxPDFFileBytes     int64 = 48 * conv.MiB
	defaultMaxTextFileBytes    int64 = 2 * conv.MiB

	defaultDocumentExtensions       = ".doc,.docx,.docm,.odt,.epub,.rtf"
	defaultImageExtensions          = "nil"
	defaultSupportedImageExtensions = ".jpg,.jpeg,.png,.webp,.tif,.tiff"
	defaultVideoExtensions          = "nil"
	defaultVisualExtensions         = ".pdf," + defaultSupportedImageExtensions

	defaultVerbose               = false
	defaultSkipFlattener         = false
	defaultSkipTARFlattener      = IsWindows
	defaultSkipFilter            = true
	defaultSkipDocumentConverter = false
	defaultSkipImageConverter    = false
	defaultSkipVideoExtractor    = false
	defaultSkipVisualMerger      = false
	defaultSkipTextMerger        = false
)

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
