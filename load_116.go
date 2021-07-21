// +build go1.16

package lunar

import (
	"embed"
	"io"
)

//go:embed files
var files embed.FS

func init() {
	loadFileFunc = func(name string) (io.ReadCloser, error) {
		return files.Open("files/" + name)
	}
}
