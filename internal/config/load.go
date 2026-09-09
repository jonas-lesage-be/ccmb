package config

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var (
	errSourceDirRequired = errors.New("source-dir is required")
	errTargetDirRequired = errors.New("target-dir is required")
)

// NewViper returns a Viper instance configured with the application defaults.
func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	v.SetDefault("source_dir", defaultSourceDir)
	v.SetDefault("target_dir", defaultTargetDir)
	v.SetDefault("target_dir_permissions", defaultTargetDirPermissions)
	v.SetDefault("target_file_permissions", defaultTargetFilePermissions)
	v.SetDefault("text_file_permissions", defaultTextFilePermissions)

	v.SetDefault("max_flattener_workers", workerCountFromPct(defaultMaxFlattenerWorkers))
	v.SetDefault("max_document_workers", workerCountFromPct(defaultMaxDocumentWorkers))
	v.SetDefault("max_image_workers", workerCountFromPct(defaultMaxImageWorkers))
	v.SetDefault("max_video_workers", workerCountFromPct(defaultMaxVideoWorkers))
	v.SetDefault("max_visual_merge_workers", workerCountFromPct(defaultMaxVisualMergeWorkers))
	v.SetDefault("image_conversion_timeout", defaultImageConversionTimeout)
	v.SetDefault("video_extract_timeout", defaultVideoExtractTimeout)

	v.SetDefault("svg_canvas_resolution", defaultSvgCanvasResolution)
	v.SetDefault("video_fps", defaultVideoFps)

	v.SetDefault("est_file_count", defaultEstFileCount)
	v.SetDefault("flat_path_delimiter", defaultFlatPathDelimiter)
	v.SetDefault("escaped_delimiter", defaultFlatPathDelimiter+defaultFlatPathDelimiter)
	v.SetDefault("decode_placeholder", defaultDecodePlaceholder)
	v.SetDefault("visual_part_prefix", defaultVisualPartPrefix)
	v.SetDefault("text_part_prefix", defaultTextPartPrefix)

	v.SetDefault("max_archive_file_bytes", defaultMaxArchiveFileBytes)
	v.SetDefault("max_pdf_file_bytes", defaultMaxPDFFileBytes)
	v.SetDefault("max_text_file_bytes", defaultMaxTextFileBytes)

	v.SetDefault("disable_in_memory_pdf", defaultDisableInMemoryPDF)

	v.SetDefault("filter_directories", parseBoolMap(defaultFilterDirectories))

	v.SetDefault("filter_extensions", parseBoolMap(defaultFilterExtensions))
	v.SetDefault("document_extensions", parseBoolMap(defaultDocumentExtensions))
	v.SetDefault("image_extensions", parseBoolMap(defaultImageExtensions))
	v.SetDefault("supported_image_extensions", parseBoolMap(defaultSupportedImageExtensions))
	v.SetDefault("video_extensions", parseBoolMap(defaultVideoExtensions))
	v.SetDefault("visual_extensions", parseBoolMap(defaultVisualExtensions))

	v.SetDefault("verbose", defaultVerbose)
	v.SetDefault("skip_flattener", defaultSkipFlattener)
	v.SetDefault("skip_document_converter", defaultSkipDocumentConverter)
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

	cfg.MaxFlattenerWorkers = adjustWorkerCount(cfg.MaxFlattenerWorkers)
	cfg.MaxDocumentWorkers = adjustWorkerCount(cfg.MaxDocumentWorkers)
	cfg.MaxImageWorkers = adjustWorkerCount(cfg.MaxImageWorkers)
	cfg.MaxVideoWorkers = adjustWorkerCount(cfg.MaxVideoWorkers)
	cfg.MaxVisualMergeWorkers = adjustWorkerCount(cfg.MaxVisualMergeWorkers)

	cfg.FilterExtensions = addDotPrefix(cfg.FilterExtensions)
	cfg.DocumentExtensions = addDotPrefix(cfg.DocumentExtensions)
	cfg.ImageExtensions = addDotPrefix(cfg.ImageExtensions)
	cfg.SupportedImageExtensions = addDotPrefix(cfg.SupportedImageExtensions)
	cfg.VideoExtensions = addDotPrefix(cfg.VideoExtensions)
	cfg.VisualExtensions = addDotPrefix(cfg.VisualExtensions)

	return &cfg, nil
}

func addDotPrefix(m map[string]bool) map[string]bool {
	if m == nil {
		return nil
	}

	res := make(map[string]bool, len(m))
	for key, val := range m {
		if !strings.HasPrefix(key, ".") {
			key = "." + key
		}
		res[key] = val
	}

	return res
}

func adjustWorkerCount(count int) int {
	if count < 0 {
		workers := runtime.NumCPU() + count
		return max(1, workers)
	}

	if count == 0 {
		return runtime.NumCPU()
	}

	return count
}

func workerCountFromPct(percentage float64) int {
	if percentage <= 0.0 {
		return 1
	}

	workers := int(float64(runtime.NumCPU()) * percentage)
	return max(1, workers)
}
