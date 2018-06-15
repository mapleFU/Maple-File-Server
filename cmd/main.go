package main

import (

	//"fmt"
	//"unsafe"
	"github.com/sirupsen/logrus"
	"github.com/mapleFU/TongjiFileLab/mapleFS"
)

func init()  {
	logrus.SetLevel(logrus.InfoLevel)
}

func main()  {
	mapleFS.Serve()
}
