// Package adminui embeds the built administrator web interface.
package adminui

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:generate npm run build

//go:embed all:dist
var distFiles embed.FS

// Dist returns the embedded administrator interface rooted at the dist
// directory.
func Dist() fs.FS {
	dist, err := fs.Sub(distFiles, "dist")
	if err != nil {
		panic(fmt.Sprintf("admin-ui: open embedded dist: %v", err))
	}

	return dist
}
