//go:build !dev

//go:generate go tool esbuild static/index.mjs --bundle --minify --outdir=static/

package public

import (
	"embed"
	"net/http"
)

var (
	//go:embed index.html static/index.js
	e embed.FS

	FS = http.FS(e)
)
