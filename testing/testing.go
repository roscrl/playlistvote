package testing

import (
	"log"
	"os"
	"path"
	"runtime"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")

	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
}
