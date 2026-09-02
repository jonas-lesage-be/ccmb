# Context Combiner (`ccmb`)

Context Combiner (`ccmb`) is a command-line tool that flattens, filters, converts, and merges files from a source directory into a target directory. It supports various file types, including image, video, and text files, and allows for size-constrained merging of visual and text data.

The tool can be configured via command-line flags, environment variables, or a configuration file in JSON, TOML, or YAML format.
Flags and environment variables take precedence over configuration file values.

## Running it

You must have the following external dependencies installed:

```sh
# LibreOffice
winget install -e --id TheDocumentFoundation.LibreOffice
# Pandoc
winget install -e --id JohnMacFarlane.Pandoc
# FFmpeg
winget install -e --id Gyan.FFmpeg
```

Ensure you have a working installation of **Go** and the `golangci-lint` tool for linting. After installing the dependencies, run the following commands to tidy up dependencies and execute the application:

```bash
go mod tidy
go run . --source-dir "./source-dir" --target-dir "./_ccmb-output"
```

To compile a minimized, highly optimized production binary stripped of debugging symbol tables:

```bash
go build -ldflags="-s -w" .
```

## Configuration

The application is configured through command-line flags. These flags are parsed dynamically at execution start and maps values using `spf13/viper` environment and configuration schemas.

### Core CLI Flags

- **`-s, --source-dir`**: The target directory containing raw input asset bundles or zip containers.
- **`-t, --target-dir`**: Destination folder where the size-bounded PDF books and textbooks are compiled.
- **`-c, --config`**: Optional path to a JSON, TOML, or YAML file to override pipeline variable allocations.
- **`--skip-flattener`**: Skips the initial nested zip/tar decompression stages.
- **`--skip-document-converter`**: Bypasses rendering document layouts (`.docx`, `.xlsx`) via external engines.
- **`--skip-image-converter`**: Prevents static vector re-scaling and canvas preprocessing filters.
- **`--skip-video-extractor`**: Disables parallel multi-threaded GPU video extraction completely.
- **`--skip-visual-merger`**: Prevents assembling watermarked PDF album batches.
- **`--skip-text-merger`**: Prevents compiling concatenated textbooks from source logs.
- **`--video-fps`**: Frame extraction frequency target slice per second (default: `1.0`).
- **`-v, --verbose`**: Activates detailed micro-operation debug level logging tracing outputs.

## Development

### Prerequisites

To set up the development environment, ensure you have the following installed:

- **Go**: Version greater than or equal to the required version in the `go.mod` file.
- **golangci-lint**: The linting tool used to enforce strict code quality standards.

#### Recommended Extensions

If you use VS Code or any derivative code editor (**Antigravity IDE**, **Cursor**, **Windsurf**, etc.), install the recommended extensions at `.vscode/extensions.json`.

### Code Quality & Formatting

This project enforces strict code quality standards using modern Go tools to ensure consistent, readable, and secure code:

- **Formatters**: Automated code styling is handled via `gofmt`, `gofumpt`, `golines`, and `gci`. This setup enforces a strict 100-character line length limit, implements enhanced formatting rules, and eliminates code inconsistencies. `gci` ensures that import statements are deterministically grouped, ordered, and structured to clearly separate Go standard library packages, external third-party dependencies, and internal `ccmb` modules.
- **Linters**: Comprehensive static analysis checks are executed to maintain optimal code health, enforce idiomatic Go design patterns, and prevent architectural anti-patterns. The linting engine analyzes the abstract syntax tree to flag structural decay, package dependency violations, and poorly abstracted interfaces that harm long-term maintainability. It continuously scans the codebase for performance bottlenecks, concurrency race hazards, and general code smell. Additionally, it runs security-focused checks to guarantee safe HTTP request processing, robust handling of file permissions, and proper mitigation against common web vulnerabilities.

Ensure your local environment is configured with these project standards before submitting code.

## Layout

```text
├── doc.go                     # Documentation for the main application entry point.
├── main.go                    # Bootstrap entry point for the ccmb command-line utility.
└── internal/                  # High-performance processing modules.
    ├── cli/                   # Binds Cobra commands and maps Viper flag schemas.
    ├── config/                # Loads configuration, environment variables, and directory filters.
    ├── document/              # Document conversion that converts unsupported formats.
    ├── image/                 # Image conversion that converts unsupported formats.
    ├── media/                 # Low-level type classification and metadata extraction.
    ├── pathsafe/              # Sanitizes path segments to prevent directory traversal attacks.
    ├── pipeline/              # Orchestrates the sequential execution of the processing pipeline.
    ├── textmerge/             # Combines text streams into size-constrained text files.
    ├── video/                 # Extracts image sequences from video files.
    └── visualmerge/           # Visual content merger that converts images and videos into PDFs.
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
