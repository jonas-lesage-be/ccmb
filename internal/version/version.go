package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application.
	Version = "dev"
	// Commit is the current commit hash of the application.
	Commit = "none"
	// BuildDate is the build date of the application.
	BuildDate = "unknown"
)

// Info holds the version information of the application.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

// Get returns the version information of the application.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted string representation of the version information.
func String() string {
	info := Get()
	return fmt.Sprintf("ccmb version %s\n"+
		"  Commit:      %s\n"+
		"  Built at:    %s\n"+
		"  Go version:  %s\n"+
		"  Compiler:    %s\n"+
		"  Platform:    %s",
		info.Version, info.Commit, info.BuildDate, info.GoVersion, info.Compiler, info.Platform)
}
