//go:build !windows

package config

const (
	IsWindows               = false
	defaultFilterExtensions = ".ttf,.woff,.woff2"
)
