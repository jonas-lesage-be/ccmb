package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var (
	errInvalidBoolMapType = errors.New("expected string for bool map")
	errSourceDirRequired  = errors.New("source-dir is required")
	errTargetDirRequired  = errors.New("target-dir is required")
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
			toBoolMapHook(),
		)),
	); err != nil {
		return nil, fmt.Errorf("decode configuration: %w", err)
	}

	cfg.SkipTARFlattenerExplicit = v.IsSet("skip_tar_flattener")
	if !cfg.SkipTARFlattenerExplicit {
		cfg.SkipTARFlattener = IsWindows
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

func toBoolMapHook() mapstructure.DecodeHookFuncType {
	target := reflect.TypeFor[map[string]bool]()

	return func(from, to reflect.Type, data any) (any, error) {
		if to != target {
			return data, nil
		}

		switch from.Kind() {
		case reflect.Map:
			return toBoolMap(data), nil
		case reflect.Slice, reflect.Array:
			return sliceToBoolMap(data), nil
		case reflect.String:
			return stringToBoolMap(data)
		default:
			return data, nil
		}
	}
}

func toBoolMap(data any) map[string]bool {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	res := make(map[string]bool, v.Len())
	for _, key := range v.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		val := v.MapIndex(key).Interface()

		if b, ok := val.(bool); ok {
			res[keyStr] = b
			continue
		}
		res[keyStr] = true
	}

	if len(res) == 0 {
		return nil
	}

	return res
}

func sliceToBoolMap(data any) map[string]bool {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	var res map[string]bool

	for i := range v.Len() {
		val := v.Index(i).Interface()
		if m, ok := val.(map[string]any); ok {
			if res == nil {
				res = make(map[string]bool)
			}

			mergeRawMap(res, m)
			continue
		}

		keyStr := strings.TrimSpace(fmt.Sprintf("%v", val))
		if keyStr != "" {
			if res == nil {
				res = make(map[string]bool)
			}

			res[keyStr] = true
		}
	}
	return res
}

func mergeRawMap(dest map[string]bool, src map[string]any) {
	for key, val := range src {
		if b, ok := val.(bool); ok {
			dest[key] = b
			continue
		}
		dest[key] = true
	}
}

func stringToBoolMap(data any) (map[string]bool, error) {
	str, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("%w: got %T", errInvalidBoolMapType, data)
	}

	trimmed := strings.TrimSpace(str)
	if trimmed == "" || trimmed == "null" {
		return map[string]bool(nil), nil
	}

	return parseBoolMap(trimmed), nil
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
