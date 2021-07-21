// +build !go1.16

package lunar

import (
	"io"
	"log"

	"github.com/rakyll/statik/fs"

	_ "github.com/xwjdsh/lunar/statik"
)

//go:generate go run github.com/rakyll/statik -src ./files -f -tags !go1.16

func init() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	loadFileFunc = func(fileName string) (io.ReadCloser, error) {
		return statikFS.Open("/" + fileName)
	}
}
