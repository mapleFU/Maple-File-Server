package src

import (
	"os"
	log "github.com/sirupsen/logrus"

	"github.com/abiosoft/ishell"
	"fmt"
)

func initServe() {
	var err error
	fsfd, err = os.OpenFile(FS_IMG_FILE, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Read fs image error: ", err)
	}
	log.SetLevel(log.InfoLevel)
}

func AllocTest()  {
	initServe()
	log.SetLevel(log.DebugLevel)
	i1 := ialloc()
	//fsyncINode(i1)
	i2 := ialloc()
	//fsyncINode(i2)

	log.Debug(i1.dinodeData.Size, " -- ", i1.num)
	log.Debug(i2.dinodeData.Size, " -- ", i2.num)
	//if i1.num == i2.num {
	//	log.Fatalf("Test Error")
	//}
	i3 := iget(0)
	i4 := iget(1)
	log.Debug(i3.dinodeData.Size, i3.num, i4.dinodeData.Size, i4.num)
}

func Serve()  {
	initServe()
	var rootDir inode
	var curDir *inode
	ReadRoot(&rootDir)
	log.Println(rootDir.dinodeData)
	WalkDir(&rootDir)
	curDir = &rootDir

	shell := ishell.New()
	shell.Println("Welcome to maple-xv6 fs! Press help to get information")

	shell.AddCmd(&ishell.Cmd{
		Name:"cd",
		Help:"Switch Current Dir",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("You didn't input arguments")
				return
			} else if len(c.Args) > 1 {
				c.Println("Too many args!!")
				return
			}
			dirName := []byte(c.Args[0])
			dirINum := dirlookup(curDir, dirName)
			if dirINum < 0 {
				c.Println("Dir doesn't exist!")
				return
			}
			// 获得实际的inode
			dirINode := iget(int(dirINum))
			fmt.Println(dirINode)
			if dirINode.dinodeData.FileType != FILETYPE_DIRECT {
				c.Println("The type of ", c.Args[0], " is not dir")
				return
			} else {
				curDir = dirINode
				log.Info("Current direct set to ", dirINode.num)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:"ls",
		Help:"List the directory",
		Func: func(c *ishell.Context) {
			dirs := WalkDir(curDir)
			for _, d := range dirs {
				shell.Println(d.DirName(), " ", d.INum)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name:"mkdir",
		Help:"Create a directory in the current dir",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("You didn't input arguments")
				return
			}
			for createIndex := 0; createIndex < len(c.Args); createIndex++ {
				toCreateName := []byte(c.Args[createIndex])
				ifExist := dirlookup(curDir, toCreateName)
				if ifExist > 0 {
					c.Println("the direct ", string(toCreateName), " already exists")
					return
				} else {
					MkDirWithParent(toCreateName, curDir)
				}
			}

		},
	})

	shell.Run()
}
