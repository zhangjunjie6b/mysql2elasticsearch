package service

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