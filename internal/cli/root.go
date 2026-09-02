package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"ccmb/internal/config"
	"ccmb/internal/pipeline"
)

func description() string {
	return strings.TrimSpace(`
Context Combiner (ccmb) is a command-line tool that flattens, filters, converts,
and merges files from a source directory into a target directory.
It supports various file types, including image, video, and text files,
and allows for size-constrained merging of visual and text data.

The tool can be configured via command-line flags, environment variables,
or a configuration file in JSON, TOML, or YAML format.
Flags and environment variables take precedence over configuration file values.
`)
}

func exampleUsage() string {
	return strings.TrimSpace(`
# Run ccmb with a configuration file
ccmb -c config.json

# Run ccmb with command-line flags
ccmb -s /path/to/source -t /path/to/target
`)
}

// NewRootCommand creates the ccmb command.
func NewRootCommand(ctx context.Context, v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ccmb",
		Short:   "Context Combiner: flatten, filter, convert, and merge files.",
		Long:    description(),
		Example: exampleUsage(),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config flag: %w", err)
			}

			cfg, err := config.Load(v, configFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			logLevel := slog.LevelInfo
			if cfg.Verbose {
				logLevel = slog.LevelDebug
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: logLevel,
			}))
			slog.SetDefault(logger)

			p := pipeline.NewPipeline(cfg)
			if err := p.Execute(ctx); err != nil {
				return fmt.Errorf("pipeline execution failed: %w", err)
			}

			return nil
		},
	}

	cmd.SetContext(ctx)

	if err := bindFlags(cmd, v); err != nil {
		slog.Error("failed to initialize flags", "err", err)
		os.Exit(1)
	}

	return cmd
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	flags := cmd.Flags()

	flags.StringP("config", "c", "", "Path to a JSON, TOML, or YAML configuration file")

	flags.StringP("source-dir", "s", "", "Directory to process")
	flags.StringP("target-dir", "t", "", "Directory for generated files")
	flags.String(
		"target-dir-permissions",
		"",
		"Permissions for the target directory (for example, 0o700)",
	)
	flags.String(
		"target-file-permissions",
		"",
		"Permissions for generated files (for example, 0o600)",
	)
	flags.String(
		"text-file-permissions",
		"",
		"Permissions for generated text files (for example, 0o600)",
	)

	flags.Int("max-flattener-workers", 0, "Maximum concurrent flattener workers")
	flags.Int("max-document-workers", 0, "Maximum concurrent document workers")
	flags.Int("max-image-workers", 0, "Maximum concurrent image workers")
	flags.Int("max-video-workers", 0, "Maximum concurrent video workers")
	flags.Int("max-visual-merge-workers", 0, "Maximum concurrent visual merge workers")
	flags.Duration("image-conversion-timeout", 0, "Maximum duration for an image conversion")
	flags.Duration("video-extract-timeout", 0, "Maximum duration for video frame extraction")

	flags.Float64("svg-canvas-resolution", 0, "Resolution for SVG canvas in pixels per millimeter")
	flags.Float64("video-fps", 0, "Frames per second for video frame extraction")

	flags.Int("est-file-count", 0, "Estimated number of input files")
	flags.String("flat-path-delimiter", "", "Delimiter used in flattened paths")
	flags.String("escaped-delimiter", "", "Escaped flattened-path delimiter")
	flags.String("decode-placeholder", "", "Temporary delimiter decoding placeholder")
	flags.String("visual-part-prefix", "", "Prefix for visual part files")
	flags.String("text-part-prefix", "", "Prefix for text part files")

	flags.Int64("max-archive-file-bytes", 0, "Maximum archive size in bytes")
	flags.Int64("max-pdf-file-bytes", 0, "Maximum PDF size in bytes")
	flags.Int64("max-text-file-bytes", 0, "Maximum text file size in bytes")

	flags.Bool("disable-in-memory-pdf", false, "Disable in-memory PDF processing")

	flags.String("filter-directories", "", "Comma-separated directories to filter")

	flags.String("filter-extensions", "", "Comma-separated extensions to filter")
	flags.String("document-extensions", "", "Comma-separated document extensions to convert")
	flags.String("image-extensions", "", "Comma-separated image extensions to convert")
	flags.String("supported-image-extensions", "", "Comma-separated supported image extensions")
	flags.String("video-extensions", "", "Comma-separated video extensions to extract")
	flags.String("visual-extensions", "", "Comma-separated visual extensions to merge")

	flags.BoolP("verbose", "v", false, "Enable verbose logging")
	flags.Bool("skip-flattener", false, "Skip the flattener")
	flags.Bool("skip-tar-flattener", false, "Skip TAR flattener support")
	flags.Bool("skip-filter", false, "Skip the filter")
	flags.Bool("skip-document-converter", false, "Skip the document converter")
	flags.Bool("skip-image-converter", false, "Skip the image converter")
	flags.Bool("skip-video-extractor", false, "Skip the video extractor")
	flags.Bool("skip-visual-merger", false, "Skip the visual merger")
	flags.Bool("skip-text-merger", false, "Skip the text merger")

	var bindErr error
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "config" || bindErr != nil {
			return
		}

		if err := v.BindPFlag(flag.Name, flag); err != nil {
			bindErr = fmt.Errorf("failed to bind flag %s: %w", flag.Name, err)
			return
		}

		key := strings.ReplaceAll(flag.Name, "-", "_")
		if key != flag.Name {
			v.RegisterAlias(key, flag.Name)
		}
	})

	return bindErr
}
