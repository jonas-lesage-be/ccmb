package config

const (
	// IsWindows is a boolean constant that indicates whether the current operating system is Windows.
	IsWindows               = true
	tarExtensions           = ".tar,.tar.gz,.tgz,.tar.xz,.txz,.tar.zst,.tzst,.tar.bz2,.tbz2"
	defaultFilterExtensions = ".ttf,.woff,.woff2," + tarExtensions
)
