package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"ccmb/internal/conv"
)

var (
	errInvalidBoolMapType = errors.New("expected string for bool map")
	errSourceDirRequired  = errors.New("source-dir is required")
	errTargetDirRequired  = errors.New("target-dir is required")
)

// Config holds the configuration settings for the application.
type Config struct {
	SourceDir            string      `mapstructure:"source_dir"`
	TargetDir            string      `mapstructure:"target_dir"`
	TargetDirPermissions os.FileMode `mapstructure:"target_dir_permissions"`
	TextFilePermissions  os.FileMode `mapstructure:"text_file_permissions"`

	MaxFlattenerWorkers    int           `mapstructure:"max_flattener_workers"`
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
	ImageExtensions          map[string]bool `mapstructure:"image_extensions"`
	SupportedImageExtensions map[string]bool `mapstructure:"supported_image_extensions"`
	VideoExtensions          map[string]bool `mapstructure:"video_extensions"`
	VisualExtensions         map[string]bool `mapstructure:"visual_extensions"`

	Verbose                  bool `mapstructure:"verbose"`
	SkipFlattener            bool `mapstructure:"skip_flattener"`
	SkipTARFlattener         bool `mapstructure:"skip_tar_flattener"`
	SkipTARFlattenerExplicit bool `mapstructure:"-"`
	SkipFilter               bool `mapstructure:"skip_filter"`
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

	defaultImageExtensions          = "nil"
	defaultSupportedImageExtensions = ".jpg,.jpeg,.png,.webp,.tif,.tiff"
	defaultVideoExtensions          = "nil"
	defaultVisualExtensions         = ".pdf," + defaultSupportedImageExtensions

	defaultVerbose            = false
	defaultSkipFlattener      = false
	defaultSkipTARFlattener   = isWindows
	defaultSkipFilter         = false
	defaultSkipImageConverter = false
	defaultSkipVideoExtractor = false
	defaultSkipVisualMerger   = false
	defaultSkipTextMerger     = false
)

// NewViper returns a Viper instance configured with the application defaults.
func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetDefault("source_dir", defaultSourceDir)
	v.SetDefault("target_dir", defaultTargetDir)
	v.SetDefault("target_dir_permissions", defaultTargetDirPermissions)
	v.SetDefault("text_file_permissions", defaultTextFilePermissions)

	v.SetDefault("max_flattener_workers", defaultMaxFlattenerWorkers)
	v.SetDefault("max_image_workers", defaultMaxImageWorkers)
	v.SetDefault("max_video_workers", defaultMaxVideoWorkers)
	v.SetDefault("image_conversion_timeout", defaultImageConversionTimeout)
	v.SetDefault("video_extract_timeout", defaultVideoExtractTimeout)
	v.SetDefault("svg_canvas_resolution", defaultSvgCanvasResolution)

	v.SetDefault("est_file_count", defaultEstFileCount)
	v.SetDefault("flat_path_delimiter", defaultFlatPathDelimiter)
	v.SetDefault("escaped_delimiter", defaultFlatPathDelimiter+defaultFlatPathDelimiter)
	v.SetDefault("decode_placeholder", defaultDecodePlaceholder)

	v.SetDefault("max_archive_file_bytes", defaultMaxArchiveFileBytes)
	v.SetDefault("max_pdf_file_bytes", defaultMaxPDFFileBytes)
	v.SetDefault("max_text_file_bytes", defaultMaxTextFileBytes)

	v.SetDefault("filter_extensions", parseBoolMap(defaultFilterExtensions))
	v.SetDefault("image_extensions", parseBoolMap(defaultImageExtensions))
	v.SetDefault("supported_image_extensions", parseBoolMap(defaultSupportedImageExtensions))
	v.SetDefault("video_extensions", parseBoolMap(defaultVideoExtensions))
	v.SetDefault("visual_extensions", parseBoolMap(defaultVisualExtensions))

	v.SetDefault("verbose", defaultVerbose)
	v.SetDefault("skip_flattener", defaultSkipFlattener)
	v.SetDefault("skip_filter", defaultSkipFilter)
	v.SetDefault("skip_image_converter", defaultSkipImageConverter)
	v.SetDefault("skip_video_extractor", defaultSkipVideoExtractor)
	v.SetDefault("skip_visual_merger", defaultSkipVisualMerger)
	v.SetDefault("skip_text_merger", defaultSkipTextMerger)

	return v
}

// Load reads an optional config JSON, TOML, or YAML file.
// Flags and environment variables take precedence over file values.
func Load(v *viper.Viper, configFile string) (*Config, error) {
	if configFile != "" {
		v.SetConfigFile(configFile)

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file %q: %w", configFile, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(
		&cfg,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToBoolMapHook(),
		)),
	); err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}

	cfg.SkipTARFlattenerExplicit = v.IsSet("skip_tar_flattener")
	if !cfg.SkipTARFlattenerExplicit {
		cfg.SkipTARFlattener = isWindows
	}

	if cfg.SourceDir == "" {
		return nil, fmt.Errorf(
			"%w: must be set via flag, env, or config file",
			errSourceDirRequired,
		)
	}
	if cfg.TargetDir == "" {
		return nil, fmt.Errorf(
			"%w: must be set via flag, env, or config file",
			errTargetDirRequired,
		)
	}

	return &cfg, nil
}

func parseBoolMap(value string) map[string]bool {
	if value == "nil" {
		return nil
	}

	set := make(map[string]bool)
	for entry := range strings.SplitSeq(value, ",") {
		if entry = strings.TrimSpace(entry); entry != "" {
			set[entry] = true
		}
	}

	return set
}

func stringToBoolMapHook() mapstructure.DecodeHookFuncType {
	target := reflect.TypeFor[map[string]bool]()

	return func(from, to reflect.Type, data any) (any, error) {
		if from.Kind() == reflect.String && to == target {
			str, ok := data.(string)
			if !ok {
				return nil, fmt.Errorf("%w: got %T", errInvalidBoolMapType, data)
			}
			return parseBoolMap(str), nil
		}

		return data, nil
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
