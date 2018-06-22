package mapleFS

import (
	//"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"
)

func Serve() {
	InitServe()
	var rootDir INode
	var curDir *INode
	ReadRoot(&rootDir)
	log.Println(rootDir.dinodeData)

	curDir = &rootDir
	WalkDir(curDir)
	CreateFile(curDir, []byte("真名实姓"))
	fileDir := IGet(Dirlookup(curDir, []byte("真名实姓")))
	EditFile(fileDir, []byte("在很久很久以前的魔法时代，任何一位谨慎的巫师都把自己的真名实姓看作最值得珍视的密藏，同时也是对自己生命的最大威胁。因为——故事里都这么说——一旦巫师的对头掌握他的真名实姓，随便用哪种人人皆知的普通魔法都能杀死他，或是使他成为自己的奴隶，无论这位巫师的魔力多么高强，而他的对头又是多么虚弱、笨拙。\n世易时移，我们人类成长了，进入理智时代，随之而来的是第一次、第二次工业革命。魔法时代的陈腐观念被抛弃了。"))
	log.Info(string(ReadFile(fileDir)))
	//shell := ishell.New()
	//shell.Println("Welcome to maple-xv6 fs! Press help to get information")
	//
	//shell.AddCmd(&ishell.Cmd{
	//	Name: "cd",
	//	Help: "Switch Current Dir",
	//	Func: func(c *ishell.Context) {
	//		if len(c.Args) == 0 {
	//			c.Println("You didn't input arguments")
	//			return
	//		} else if len(c.Args) > 1 {
	//			c.Println("Too many args!!")
	//			return
	//		}
	//		dirName := []byte(c.Args[0])
	//		dirINum := Dirlookup(curDir, dirName)
	//		if dirINum < 0 {
	//			c.Println("Dir doesn't exist!")
	//			return
	//		}
	//		// 获得实际的inode
	//		dirINode := IGet(int(dirINum))
	//
	//		if dirINode.dinodeData.FileType != FILETYPE_DIRECT {
	//			c.Println("The type of ", c.Args[0], " is not dir")
	//			return
	//		} else {
	//			curDir = dirINode
	//			log.Info("Current direct set to ", dirINode.num)
	//		}
	//	},
	//})
	//
	//shell.AddCmd(&ishell.Cmd{
	//	Name: "ls",
	//	Help: "List the directory",
	//	Func: func(c *ishell.Context) {
	//		log.Info("Ready to ls")
	//		dirs := WalkDir(curDir)
	//		for _, d := range dirs {
	//			shell.Println(d.DirName(), " ", d.INum)
	//		}
	//	},
	//})
	//
	//shell.AddCmd(&ishell.Cmd{
	//	Name: "mkdir",
	//	Help: "Create a directory in the current dir",
	//	Func: func(c *ishell.Context) {
	//		if len(c.Args) == 0 {
	//			c.Println("You didn't input arguments")
	//			return
	//		}
	//		for createIndex := 0; createIndex < len(c.Args); createIndex++ {
	//			toCreateName := []byte(c.Args[createIndex])
	//			ifExist := Dirlookup(curDir, toCreateName)
	//			if ifExist > 0 {
	//				c.Println("the direct ", string(toCreateName), " already exists")
	//				return
	//			} else {
	//				MkDirWithParent(toCreateName, curDir)
	//			}
	//		}
	//
	//	},
	//})
	//
	//shell.AddCmd(&ishell.Cmd{
	//	Name: "rmdir",
	//	Help: "Create an empty directory in the current dir",
	//	Func: func(c *ishell.Context) {
	//		if len(c.Args) == 0 {
	//			c.Println("You didn't input arguments")
	//			return
	//		}
	//
	//		for createIndex := 0; createIndex < len(c.Args); createIndex++ {
	//			toCreateName := []byte(c.Args[createIndex])
	//
	//			ifExist := Dirlookup(curDir, toCreateName)
	//			log.Info(ifExist)
	//			if ifExist <= 0 {
	//				log.Info("the direct ", string(toCreateName), " non exists")
	//				return
	//			} else {
	//				log.Info("dirunlink ", string(toCreateName))
	//				dirunlink(curDir, uint16(ifExist))
	//				//MkDirWithParent(toCreateName, curDir)
	//			}
	//		}
	//
	//	},
	//})
	//
	//shell.Run()
}
