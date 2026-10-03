# Context Combiner (`ccmb`)

Context Combiner (`ccmb`) is a command-line tool that flattens, filters, converts, and merges files from a source directory into a target directory. It supports various file types, including image, video, and text files, and allows for size-constrained merging of visual and text data.

The tool can be configured via command-line flags, environment variables, or a configuration file in JSON, TOML, or YAML format.
Flags and environment variables take precedence over configuration file values.

## Installation

Choose your preferred package manager to install `ccmb`:

<details>
<summary>Docker (Docker Hub / GitHub Container Registry)</summary>

### Docker Hub

```bash
docker pull contextcombiner/ccmb
```

### GitHub Container Registry

```bash
docker pull ghcr.io/jonas-lesage-be/ccmb
```

_Note: The Docker image includes all recommended external dependencies (LibreOffice, Pandoc, FFmpeg) to unlock the full feature set of `ccmb`._

</details>

<details>
<summary>Windows (WinGet / Chocolatey / Scoop)</summary>

### WinGet

```powershell
winget install -e --id ccmb.ccmb
```

### Chocolatey

```powershell
choco install ccmb
```

### Scoop

```powershell
scoop bucket add jonas-lesage-be https://github.com/jonas-lesage-be/scoop-bucket
scoop install ccmb
```

</details>

<details>
<summary>macOS (Homebrew)</summary>

### Homebrew

```bash
brew tap jonas-lesage-be/homebrew-tap
brew install --cask ccmb
```

</details>

<details>
<summary>Linux (APT / DNF / Pacman / Zypper / APK)</summary>

Go to the [releases](https://github.com/jonas-lesage-be/ccmb/releases) page and download the package file for your distribution. Then install it using the appropriate package manager:

- **Debian / Ubuntu:** Download the `.deb` file and install with `sudo dpkg -i <file>.deb`
- **Fedora / RHEL:** Download the `.rpm` file and install with `sudo dnf install <file>.rpm`
- **Arch Linux:** Download the `.pkg.tar.zst` file and install with `sudo pacman -U <file>.pkg.tar.zst`
- **openSUSE:** Download the `.rpm` file and install with `sudo zypper install --allow-unsigned-rpm <file>.rpm`
- **Alpine Linux:** Download the `.apk` file and install with `sudo apk add --allow-untrusted <file>.apk`

</details>

## Recommended external dependencies

The following external dependencies are recommended to unlock the full feature set of `ccmb`:

- **LibreOffice**: For document conversion and manipulation.
- **Pandoc**: For converting between various document formats.
- **FFmpeg**: For image and video processing.

You can install these dependencies using your preferred package manager:

<details>
<summary>Windows (WinGet / Chocolatey / Scoop)</summary>

### WinGet

```powershell
winget install -e --id TheDocumentFoundation.LibreOffice && winget install -e --id JohnMacFarlane.Pandoc && winget install -e --id Gyan.FFmpeg
```

### Chocolatey

```powershell
choco install libreoffice pandoc ffmpeg -y
```

### Scoop

```powershell
scoop bucket add extras && scoop install libreoffice pandoc ffmpeg
```

_Note: The `extras` bucket is required for LibreOffice._

</details>

<details>
<summary>macOS (Homebrew)</summary>

### Homebrew

```bash
brew install --cask libreoffice && brew install pandoc ffmpeg
```

</details>

<details>
<summary>Linux (APT / DNF / Pacman / Zypper / APK)</summary>

### Ubuntu / Debian (APT)

```bash
sudo apt update && sudo apt install -y libreoffice pandoc ffmpeg
```

### Fedora / RHEL (DNF)

```bash
sudo dnf install -y libreoffice pandoc ffmpeg
```

### Arch Linux (Pacman)

```bash
sudo pacman -Syu --noconfirm libreoffice-fresh pandoc-cli ffmpeg
```

### openSUSE (Zypper)

```bash
sudo zypper install -y libreoffice pandoc ffmpeg
```

### Alpine Linux (APK)

```bash
sudo apk add libreoffice pandoc-cli ffmpeg
```

</details>

## Usage

```bash
ccmb -s "./source-dir" -t "./_ccmb-output"
```

## Configuration

The application is configured via command-line flags that are parsed at startup.

### Core CLI flags

| Flag                        | Description                                                 | Default            |
| :-------------------------- | :---------------------------------------------------------- | :----------------- |
| `-c, --config <path>`       | Path to a JSON, TOML, or YAML configuration file.           | `""`               |
| `-s, --source-dir <path>`   | Source directory containing raw input files.                | `"."`              |
| `-t, --target-dir <path>`   | Target directory to store output files.                     | `"./_ccmb-output"` |
| `--video-fps <float>`       | Frames per second for video frame extraction.               | `1.0`              |
| `--skip-flattener`          | Skips the archive extraction and directory flattening step. | `false`            |
| `--skip-document-converter` | Skips the document conversion step (LibreOffice/Pandoc).    | `false`            |
| `--skip-image-converter`    | Skips the image conversion step.                            | `false`            |
| `--skip-video-extractor`    | Skips the video frame extraction step (FFmpeg).             | `false`            |
| `--skip-visual-merger`      | Skips the visual content merging step.                      | `false`            |
| `--skip-text-merger`        | Skips the text stream merging step.                         | `false`            |
| `--verbose`                 | Enable verbose debug logging output.                        | `false`            |

## Development

### Prerequisites

To set up the development environment, ensure you have the following installed:

- **Go**: Version greater than or equal to the required version in the `go.mod` file.
- **golangci-lint**: The linting tool used to enforce strict code quality standards.

#### Recommended extensions

If you use VS Code or any derivative code editor (**Antigravity IDE**, **Cursor**, **Windsurf**, etc.), install the recommended extensions at `.vscode/extensions.json`.

### Code quality and formatting

This project enforces strict code quality standards using modern Go tools to ensure consistent, readable, and secure code:

- **Formatters**: Automated code styling is handled via `gofmt`, `gofumpt`, `golines`, and `gci`. This setup enforces a strict 100-character line length limit, implements enhanced formatting rules, and eliminates code inconsistencies. `gci` ensures that import statements are deterministically grouped, ordered, and structured to clearly separate Go standard library packages, external third-party dependencies, and internal `ccmb` modules.
- **Linters**: Comprehensive static analysis checks are executed to maintain optimal code health, enforce idiomatic Go design patterns, and prevent architectural anti-patterns. The linting engine analyzes the abstract syntax tree to flag structural decay, package dependency violations, and poorly abstracted interfaces that harm long-term maintainability. It continuously scans the codebase for performance bottlenecks, concurrency race hazards, and general code smell. Additionally, it runs security-focused checks to guarantee safe HTTP request processing, robust handling of file permissions, and proper mitigation against common web vulnerabilities.

Ensure your local environment is configured with these project standards before submitting code.

## Project layout

```text
├── doc.go                     # Documentation for the main application entry point.
├── main.go                    # Bootstrap entry point for the ccmb command-line utility.
└── internal/                  # High-performance processing modules.
    ├── cli/                   # Binds Cobra commands and maps Viper flag schemas.
    ├── config/                # Loads configuration.
    ├── document/              # Document conversion that converts unsupported formats.
    ├── flattener/             # Extracts archive files and flattens directory structures.
    ├── image/                 # Image conversion that converts unsupported formats.
    ├── media/                 # Low-level type classification and metadata extraction.
    ├── pathsafe/              # Sanitizes path segments to prevent directory traversal attacks.
    ├── pipeline/              # Orchestrates the sequential execution of the processing pipeline.
    ├── textmerge/             # Combines text streams into size-constrained text files.
    ├── units/                 # Defines file size units.
    ├── version/               # Provides version information.
    ├── video/                 # Extracts image sequences from video files.
    └── visualmerge/           # Visual content merger that converts images and videos into PDFs.
```

## Contributing

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/name`).
3. Ensure all code passes `golangci-lint`.
4. Open a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
