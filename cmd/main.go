package main

import (
	"github.com/mapleFU/TongjiFileLab/src"

	//"fmt"
	//"unsafe"
	"github.com/sirupsen/logrus"
)

func init()  {
	logrus.SetLevel(logrus.InfoLevel)
}

func main()  {
	src.Serve()
}
