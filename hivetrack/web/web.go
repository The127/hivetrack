// Package web holds the embedded frontend assets.
// The dist/ directory is populated by `just ui-build` before compiling.
package web

import "embed"

//go:embed all:dist
var Assets embed.FS
