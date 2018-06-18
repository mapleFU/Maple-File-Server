package mapleFS

import (
	"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"
)

func Serve() {
	initServe()
	var rootDir INode
	var curDir *INode
	ReadRoot(&rootDir)
	log.Println(rootDir.dinodeData)
	WalkDir(&rootDir)
	curDir = &rootDir

	shell := ishell.New()
	shell.Println("Welcome to maple-xv6 fs! Press help to get information")

	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "Switch Current Dir",
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
		Name: "ls",
		Help: "List the directory",
		Func: func(c *ishell.Context) {
			dirs := WalkDir(curDir)
			for _, d := range dirs {
				shell.Println(d.DirName(), " ", d.INum)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "mkdir",
		Help: "Create a directory in the current dir",
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
