//go:build !dev

package webspa

import "embed"

//go:embed out
var SPA embed.FS
