package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccmb/internal/config"
	"ccmb/internal/conv"
)

func TestNewRootCommand_UsesDefaultsWithoutFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	v := config.NewViper()
	cmd := NewRootCommand(ctx, v)

	// Simulate an empty terminal call where the user passes zero flags.
	flagArgs := []string{}

	// Act
	err := cmd.ParseFlags(flagArgs)
	require.NoError(t, err)

	cfg, err := config.Load(v, "")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ".", cfg.SourceDir)
	assert.Equal(t, "./_ccmb_output", cfg.TargetDir)
	assert.Equal(t, os.FileMode(0o700), cfg.TargetDirPermissions)
	assert.Equal(t, os.FileMode(0o600), cfg.TargetFilePermissions)
	assert.Equal(t, os.FileMode(0o600), cfg.TextFilePermissions)

	assert.Equal(t, 8, cfg.MaxFlattenerWorkers)
	assert.Equal(t, 8, cfg.MaxDocumentWorkers)
	assert.Equal(t, 8, cfg.MaxImageWorkers)
	assert.Equal(t, 8, cfg.MaxVideoWorkers)
	assert.Equal(t, 16, cfg.MaxVisualMergeWorkers)
	assert.Equal(t, 15*time.Second, cfg.ImageConversionTimeout)
	assert.Equal(t, 1*time.Minute, cfg.VideoExtractTimeout)

	assert.InDelta(t, 300.0/25.4, cfg.SvgCanvasResolution, 0.000001)
	assert.InDelta(t, 1.0, cfg.VideoFps, 0.000001)

	assert.Equal(t, 128, cfg.EstFileCount)
	assert.Equal(t, "--", cfg.FlatPathDelimiter)
	assert.Equal(t, "----", cfg.EscapedDelimiter)
	assert.Equal(t, "\x00", cfg.DecodePlaceholder)
	assert.Equal(t, "visual-part-", cfg.VisualPartPrefix)
	assert.Equal(t, "text-part-", cfg.TextPartPrefix)

	assert.Equal(t, int64(4*conv.GiB), cfg.MaxArchiveFileBytes)
	assert.Equal(t, int64(48*conv.MiB), cfg.MaxPDFFileBytes)
	assert.Equal(t, int64(2*conv.MiB), cfg.MaxTextFileBytes)

	assert.False(t, cfg.DisableInMemoryPDF)

	assert.Nil(t, cfg.ImageExtensions)
	assert.Nil(t, cfg.VideoExtensions)

	expectedDocumentExtensions := []string{
		".csv",
		".odt",
		".ods",
		".odp",
		".epub",
		".rtf",
		".doc",
		".docx",
		".docm",
		".xls",
		".xlsx",
		".xlsm",
		".ppt",
		".pptx",
		".pptm",
	}
	for _, ext := range expectedDocumentExtensions {
		assert.True(t, cfg.DocumentExtensions[ext])
	}

	expectedFilterExtensions := []string{".ttf", ".woff", ".woff2"}
	if config.IsWindows {
		expectedFilterExtensions = append(
			expectedFilterExtensions,
			".tar",
			".tar.gz",
			".tgz",
			".tar.xz",
			".txz",
			".tar.zst",
			".tzst",
			".tar.bz2",
			".tbz2",
		)
	}
	for _, ext := range expectedFilterExtensions {
		assert.True(t, cfg.FilterExtensions[ext])
	}

	expectedSupportedExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff"}
	for _, ext := range expectedSupportedExtensions {
		assert.True(t, cfg.SupportedImageExtensions[ext])
		assert.True(t, cfg.VisualExtensions[ext])
	}
	assert.True(t, cfg.VisualExtensions[".pdf"])

	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.SkipFlattener)
	assert.False(t, cfg.SkipImageConverter)
	assert.False(t, cfg.SkipVideoExtractor)
	assert.False(t, cfg.SkipVisualMerger)
	assert.False(t, cfg.SkipTextMerger)
}

func TestNewRootCommand_FlagsOverrideDefaults(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	v := config.NewViper()
	cmd := NewRootCommand(ctx, v)

	// Simulate user terminal inputs with overrides for every single field.
	flagArgs := []string{
		"--source-dir", "override",
		"--target-dir", "overridden_target",
		"--target-dir-permissions", "0o770",
		"--target-file-permissions", "0o660",
		"--text-file-permissions", "0o660",

		"--max-flattener-workers", "10",
		"--max-document-workers", "12",
		"--max-image-workers", "20",
		"--max-video-workers", "15",
		"--max-visual-merge-workers", "8",
		"--image-conversion-timeout", "20s",
		"--video-extract-timeout", "2m",

		"--svg-canvas-resolution", "600.0",
		"--video-fps", "24.0",

		"--est-file-count", "512",
		"--flat-path-delimiter", "__",
		"--escaped-delimiter", "____",
		"--decode-placeholder", "\x00\x00",
		"--visual-part-prefix", "vis-part-",
		"--text-part-prefix", "txt-part-",

		"--max-archive-file-bytes", "2147483648", // 2 * conv.GiB
		"--max-pdf-file-bytes", "33554432", // 32 * conv.MiB
		"--max-text-file-bytes", "1048576", // 1 * conv.MiB

		"--disable-in-memory-pdf=true",

		"--filter-extensions", "val1,val2",
		"--document-extensions", "doc1,doc2",
		"--image-extensions", "img1,img2",
		"--supported-image-extensions", "s1,s2",
		"--video-extensions", "v1,v2",
		"--visual-extensions", "v3,v4",

		"--verbose=true",
		"--skip-flattener=true",
		"--skip-tar-flattener=false",
		"--skip-document-converter=true",
		"--skip-image-converter=true",
		"--skip-video-extractor=true",
		"--skip-visual-merger=true",
		"--skip-text-merger=true",
	}

	// Act
	err := cmd.ParseFlags(flagArgs)
	require.NoError(t, err)

	cfg, err := config.Load(v, "")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "override", cfg.SourceDir)
	assert.Equal(t, "overridden_target", cfg.TargetDir)
	assert.Equal(t, os.FileMode(0o770), cfg.TargetDirPermissions)
	assert.Equal(t, os.FileMode(0o660), cfg.TargetFilePermissions)
	assert.Equal(t, os.FileMode(0o660), cfg.TextFilePermissions)

	assert.Equal(t, 10, cfg.MaxFlattenerWorkers)
	assert.Equal(t, 12, cfg.MaxDocumentWorkers)
	assert.Equal(t, 20, cfg.MaxImageWorkers)
	assert.Equal(t, 15, cfg.MaxVideoWorkers)
	assert.Equal(t, 8, cfg.MaxVisualMergeWorkers)
	assert.Equal(t, 20*time.Second, cfg.ImageConversionTimeout)
	assert.Equal(t, 2*time.Minute, cfg.VideoExtractTimeout)

	assert.InDelta(t, 600.0, cfg.SvgCanvasResolution, 0.000001)
	assert.InDelta(t, 24.0, cfg.VideoFps, 0.000001)

	assert.Equal(t, 512, cfg.EstFileCount)
	assert.Equal(t, "__", cfg.FlatPathDelimiter)
	assert.Equal(t, "____", cfg.EscapedDelimiter)
	assert.Equal(t, "\x00\x00", cfg.DecodePlaceholder)
	assert.Equal(t, "vis-part-", cfg.VisualPartPrefix)
	assert.Equal(t, "txt-part-", cfg.TextPartPrefix)

	assert.Equal(t, int64(2*conv.GiB), cfg.MaxArchiveFileBytes)
	assert.Equal(t, int64(32*conv.MiB), cfg.MaxPDFFileBytes)
	assert.Equal(t, int64(1*conv.MiB), cfg.MaxTextFileBytes)

	assert.True(t, cfg.DisableInMemoryPDF)

	assert.True(t, cfg.FilterExtensions[".val1"])
	assert.True(t, cfg.FilterExtensions[".val2"])
	assert.True(t, cfg.DocumentExtensions[".doc1"])
	assert.True(t, cfg.DocumentExtensions[".doc2"])
	assert.True(t, cfg.ImageExtensions[".img1"])
	assert.True(t, cfg.SupportedImageExtensions[".s1"])
	assert.True(t, cfg.VideoExtensions[".v1"])
	assert.True(t, cfg.VisualExtensions[".v3"])

	assert.True(t, cfg.Verbose)
	assert.True(t, cfg.SkipFlattener)
	assert.False(t, cfg.SkipTARFlattener)
	assert.True(t, cfg.SkipImageConverter)
	assert.True(t, cfg.SkipVideoExtractor)
	assert.True(t, cfg.SkipVisualMerger)
	assert.True(t, cfg.SkipTextMerger)
}
