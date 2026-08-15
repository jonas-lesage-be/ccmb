package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccmb/internal/conv"
)

func TestLoad_WithConfigFiles(t *testing.T) {
	t.Parallel()

	// Arrange
	formats := []string{
		"object/config_default.json",
		"object/config_default.toml",
		"object/config_default.yaml",
		"array/config_default.json",
		"array/config_default.toml",
		"array/config_default.yaml",
		"string/config_default.json",
		"string/config_default.toml",
		"string/config_default.yaml",
	}

	for _, filename := range formats {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			// Arrange
			path := filepath.Join("testdata", filename)
			v := NewViper()

			// Act
			cfg, err := Load(v, path)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, ".", cfg.SourceDir)
			assert.Equal(t, "./_ccmb_output", cfg.TargetDir)
			assert.Equal(t, os.FileMode(0o700), cfg.TargetDirPermissions)
			assert.Equal(t, os.FileMode(0o600), cfg.TextFilePermissions)

			assert.Equal(t, 8, cfg.MaxFlattenerWorkers)
			assert.Equal(t, 8, cfg.MaxImageWorkers)
			assert.Equal(t, 4, cfg.MaxVideoWorkers)
			assert.Equal(t, 15*time.Second, cfg.ImageConversionTimeout)
			assert.Equal(t, 1*time.Minute, cfg.VideoExtractTimeout)
			assert.InDelta(t, 300/25.4, cfg.SvgCanvasResolution, 0.000001)

			assert.Equal(t, 128, cfg.EstFileCount)
			assert.Equal(t, "--", cfg.FlatPathDelimiter)
			assert.Equal(t, "----", cfg.EscapedDelimiter)
			assert.Equal(t, "\x00", cfg.DecodePlaceholder)

			assert.Equal(t, int64(4294967296), cfg.MaxArchiveFileBytes)
			assert.Equal(t, int64(50331648), cfg.MaxPDFFileBytes)
			assert.Equal(t, int64(2097152), cfg.MaxTextFileBytes)

			assert.Nil(t, cfg.ImageExtensions)
			assert.Nil(t, cfg.VideoExtensions)

			expectedFilterExtensions := []string{
				".tar",
				".tar.gz",
				".tgz",
				".tar.xz",
				".txz",
				".tar.zst",
				".tzst",
				".tar.bz2",
				".tbz2",
			}
			for _, ext := range expectedFilterExtensions {
				assert.True(t, cfg.FilterExtensions[ext])
			}

			expectedDocumentExtensions := []string{
				".doc",
				".docx",
				".docm",
				".odt",
				".epub",
				".rtf",
			}
			for _, ext := range expectedDocumentExtensions {
				assert.True(t, cfg.DocumentExtensions[ext])
			}

			expectedExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff"}
			for _, ext := range expectedExtensions {
				assert.True(t, cfg.SupportedImageExtensions[ext])
				assert.True(t, cfg.VisualExtensions[ext])
			}
			assert.True(t, cfg.VisualExtensions[".pdf"])

			assert.False(t, cfg.Verbose)
			assert.False(t, cfg.SkipFlattener)
			assert.True(t, cfg.SkipTARFlattener)
			assert.True(t, cfg.SkipFilter)
			assert.False(t, cfg.SkipImageConverter)
			assert.False(t, cfg.SkipVideoExtractor)
			assert.False(t, cfg.SkipVisualMerger)
			assert.False(t, cfg.SkipTextMerger)
		})
	}
}

func TestLoad_UsesOverrides(t *testing.T) {
	t.Parallel()

	// Arrange
	v := NewViper()
	v.Set("source_dir", "override")
	v.Set("target_dir", "overridden_target")
	v.Set("target_dir_permissions", 0o770)
	v.Set("text_file_permissions", 0o660)

	v.Set("max_flattener_workers", 10)
	v.Set("max_image_workers", 20)
	v.Set("max_video_workers", 15)
	v.Set("image_conversion_timeout", "20s")
	v.Set("video_extract_timeout", "2m")
	v.Set("svg_canvas_resolution", 600.0)

	v.Set("est_file_count", 512)
	v.Set("flat_path_delimiter", "__")
	v.Set("escaped_delimiter", "____")
	v.Set("decode_placeholder", "\x00\x00")

	v.Set("max_archive_file_bytes", 2*conv.GiB)
	v.Set("max_pdf_file_bytes", 32*conv.MiB)
	v.Set("max_text_file_bytes", 1*conv.MiB)

	v.Set("filter_extensions", "val1,val2")
	v.Set("image_extensions", "img1,img2")
	v.Set("supported_image_extensions", "s1,s2")
	v.Set("video_extensions", "v1,v2")
	v.Set("visual_extensions", "v3,v4")

	v.Set("verbose", !defaultVerbose)
	v.Set("skip_flattener", !defaultSkipFlattener)
	v.Set("skip_tar_flattener", !defaultSkipTARFlattener)
	v.Set("skip_filter", !defaultSkipFilter)
	v.Set("skip_document_converter", !defaultSkipDocumentConverter)
	v.Set("skip_image_converter", !defaultSkipImageConverter)
	v.Set("skip_video_extractor", !defaultSkipVideoExtractor)
	v.Set("skip_visual_merger", !defaultSkipVisualMerger)
	v.Set("skip_text_merger", !defaultSkipTextMerger)

	// Act
	cfg, err := Load(v, "")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "override", cfg.SourceDir)
	assert.Equal(t, "overridden_target", cfg.TargetDir)
	assert.Equal(t, os.FileMode(0o770), cfg.TargetDirPermissions)
	assert.Equal(t, os.FileMode(0o660), cfg.TextFilePermissions)

	assert.Equal(t, 10, cfg.MaxFlattenerWorkers)
	assert.Equal(t, 20, cfg.MaxImageWorkers)
	assert.Equal(t, 15, cfg.MaxVideoWorkers)
	assert.Equal(t, 20*time.Second, cfg.ImageConversionTimeout)
	assert.Equal(t, 2*time.Minute, cfg.VideoExtractTimeout)
	assert.InDelta(t, 600.0, cfg.SvgCanvasResolution, 0.000001)

	assert.Equal(t, 512, cfg.EstFileCount)
	assert.Equal(t, "__", cfg.FlatPathDelimiter)
	assert.Equal(t, "____", cfg.EscapedDelimiter)
	assert.Equal(t, "\x00\x00", cfg.DecodePlaceholder)

	assert.Equal(t, int64(2*conv.GiB), cfg.MaxArchiveFileBytes)
	assert.Equal(t, int64(32*conv.MiB), cfg.MaxPDFFileBytes)
	assert.Equal(t, int64(1*conv.MiB), cfg.MaxTextFileBytes)

	assert.Contains(t, cfg.FilterExtensions, "val1")
	assert.True(t, cfg.FilterExtensions["val2"])
	assert.True(t, cfg.ImageExtensions["img1"])
	assert.True(t, cfg.SupportedImageExtensions["s1"])
	assert.True(t, cfg.VideoExtensions["v1"])
	assert.True(t, cfg.VisualExtensions["v3"])

	assert.Equal(t, !defaultVerbose, cfg.Verbose)
	assert.Equal(t, !defaultSkipFlattener, cfg.SkipFlattener)
	assert.Equal(t, !defaultSkipTARFlattener, cfg.SkipTARFlattener)
	assert.Equal(t, !defaultSkipFilter, cfg.SkipFilter)
	assert.Equal(t, !defaultSkipDocumentConverter, cfg.SkipDocumentConverter)
	assert.Equal(t, !defaultSkipImageConverter, cfg.SkipImageConverter)
	assert.Equal(t, !defaultSkipVideoExtractor, cfg.SkipVideoExtractor)
	assert.Equal(t, !defaultSkipVisualMerger, cfg.SkipVisualMerger)
	assert.Equal(t, !defaultSkipTextMerger, cfg.SkipTextMerger)
}
