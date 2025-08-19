//go:build dev

//go:generate go tool esbuild static/index.mjs --bundle --sourcemap --outdir=static/

package public

import (
	"net/http"
)

var FS = http.Dir("./public")
