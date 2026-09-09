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

Ensure you have a working installation of **Go** and the `golangci-lint` tool for formatting and linting. After installing the dependencies, run the following commands to tidy up dependencies and execute the application:

```bash
go mod tidy
go run . --source-dir "./source-dir" --target-dir "./_ccmb-output"
```

## Building from source

The easiest way to build the project is by using the provided Makefile:

```bash
make build
```

Alternatively, if you don't have `make` installed, you can compile a minimized, highly optimized production binary stripped of debugging symbol tables using the raw Go command:

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" .
```

## Configuration

The application is configured through command-line flags. These flags are parsed dynamically at execution start and maps values using `spf13/viper` environment and configuration schemas.

### Core CLI Flags

- **`-c, --config`**: Path to a JSON, TOML, or YAML configuration file.
- **`-s, --source-dir`**: Source directory containing raw input files.
- **`-t, --target-dir`**: Target directory to store output files.
- **`--video-fps`**: Frames per second for video frame extraction (default: `1.0`).
- **`--skip-flattener`**: Skips the flattener.
- **`--skip-document-converter`**: Skips the document converter.
- **`--skip-image-converter`**: Skips the image converter.
- **`--skip-video-extractor`**: Skips the video extractor.
- **`--skip-visual-merger`**: Skips the visual merger.
- **`--skip-text-merger`**: Skips the text merger.
- **`-v, --verbose`**: Enable verbose logging.

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
    ├── flattener/             # Extracts archive files and flattens directory structures.
    ├── image/                 # Image conversion that converts unsupported formats.
    ├── media/                 # Low-level type classification and metadata extraction.
    ├── pathsafe/              # Sanitizes path segments to prevent directory traversal attacks.
    ├── pipeline/              # Orchestrates the sequential execution of the processing pipeline.
    ├── textmerge/             # Combines text streams into size-constrained text files.
    ├── units/                 # Defines file size units.
    ├── video/                 # Extracts image sequences from video files.
    └── visualmerge/           # Visual content merger that converts images and videos into PDFs.
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
