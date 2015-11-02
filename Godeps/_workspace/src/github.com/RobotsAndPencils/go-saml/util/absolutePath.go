package util

import (
	"fmt"
	"path"

	"github.com/kardianos/osext"
)

func AbsolutePath(aPath string) string {
	if path.IsAbs(aPath) {
		return aPath
	}
	wd, err := osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}
	fmt.Println("Working directory", wd)
	return path.Join(wd, aPath)
}
