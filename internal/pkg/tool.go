package pkg

import (
	"io"
	"os"
)

func CopyFile(dst string, src string)  {
	//TODO 简单逻辑后续做多于判断
	srcfile,_ := os.Open(src)
	dstfile,_ := os.Create(dst)
	io.Copy(dstfile, srcfile)
}

func SliceIn(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
