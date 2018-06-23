package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mapleFU/TongjiFileLab/mapleFS"
	"net/url"
	//"log"
	//log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus"
	"strings"
)

type FileSchema struct {
	FileName string `json:"file_name"`
	INum     uint16 `json:"i_num"`
	FileType string `json:"file_type"`
}

func main() {
	r := gin.Default()
	mapleFS.InitServe()
	var rootDir mapleFS.INode

	// 名称stack
	dirNameStack := make([]string, 0)
	dirNameStack = append(dirNameStack, "root")

	mapleFS.ReadRoot(&rootDir)
	currentDir := &rootDir

	// 检查名称是否存在，如果存在折返回 false
	checkNameExists := func(context *gin.Context) (bool, string) {
		name := context.Param("name")
		if strings.Compare(name, "") == 0 {
			context.JSON(422, map[string]string{
				"error": "You lost 'name' argument in post /dirs/{name}",
			})
			return false, ""
		}
		iNum := mapleFS.Dirlookup(currentDir, []byte(name))
		if iNum != -1 {
			// already exists
			context.JSON(409, map[string]string{
				"error": fmt.Sprintf("file %s already exists in current dir", name),
			})
			return false, ""
		}
		return true, name
	}

	// 检查名称是否存在，如果不存在折返回 false
	checkNameUnExists := func(context *gin.Context) (bool, int) {

		name := context.Param("name")
		logrus.Info("Name is ", name, " in the requests.")
		if strings.Compare(name, "") == 0 {
			context.JSON(422, map[string]string{
				"error": "You lost 'name' argument in post /dirs/{name}",
			})
			return false, -1

		}
		iNum := mapleFS.Dirlookup(currentDir, []byte(name))
		if iNum == -1 {
			// already exists
			context.JSON(404, map[string]string{
				"error": fmt.Sprintf("file %s not found in current dir", name),
			})
			return false, -1
		}
		return true, iNum
	}

	// serve index page
	r.StaticFile("/index", "template/frontPage.html")
	r.StaticFile("", "template/frontPage.html")

	// list files
	r.GET("/dirs", func(context *gin.Context) {
		curFiles := mapleFS.WalkDir(currentDir)

		var fileDatas []FileSchema
		for _, dirent := range curFiles {
			fileDatas = append(fileDatas, FileSchema{
				FileName: dirent.DirName(),
				INum:     dirent.INum,
				FileType: dirent.GetType(),
			})
		}

		logrus.Info("pwd get: ", strings.Join(dirNameStack, "/"))
		context.JSON(200, map[string]interface{}{
			"data":    fileDatas,
			"current": strings.Join(dirNameStack, "/"),
		})
	})

	// read concrete file, like cat
	r.GET("/files/:name", func(context *gin.Context) {
		// watch a file
		fileName := context.Param("name")
		iNum := mapleFS.Dirlookup(currentDir, []byte(fileName))
		if iNum == -1 {
			context.JSON(404, nil)
		}
		var resultMap = map[string]string{
			"text": string(mapleFS.ReadFileFromINum(uint16(iNum))),
		}

		context.JSON(200, resultMap)
	})

	// cd
	r.GET("/cd/*name", func(context *gin.Context) {
		name := context.Param("name")
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		if strings.Compare(name, "") == 0 {
			if len(dirNameStack) != 1 {
				dirNameStack = dirNameStack[:len(dirNameStack)-1]
			}
			currentDir = mapleFS.IGet(mapleFS.Dirlookup(currentDir, []byte("..")))
			context.JSON(204, nil)
		} else {
			iNum := mapleFS.Dirlookup(currentDir, []byte(name))
			if iNum == -1 {
				context.JSON(404, map[string]string{
					"error": "file " + name + " not exist in dir " + dirNameStack[len(dirNameStack)-1],
				})
				return
			}
			if name == ".." {
				if len(dirNameStack) != 1 {
					dirNameStack = dirNameStack[:len(dirNameStack)-1]
				}
			} else {
				dirNameStack = append(dirNameStack, name)
			}
			currentDir = mapleFS.IGet(iNum)

			context.JSON(204, nil)
		}
	})

	// mkdir
	r.POST("/dirs/:name", func(context *gin.Context) {

		nonExists, name := checkNameExists(context)
		logrus.Info("Create dir ", name)
		if !nonExists {
			return
		}
		// parse arguments now
		iNode := mapleFS.MkDirWithParent([]byte(name), currentDir)
		u := url.URL{}
		u.Host = "127.0.0.1"
		u.Path = "/dirs/" + name
		context.Header("Location", u.String())
		// ?
		context.JSON(201, FileSchema{
			name,
			iNode.GetINum(),
			iNode.GetType(),
		})
	})

	r.DELETE("/files/:name", func(context *gin.Context) {
		exists, iNum := checkNameUnExists(context)
		if !exists {
			return
		}
		file := mapleFS.IGet(iNum)
		if !file.IsFile() {
			context.JSON(400, map[string]string{
				"error": "file is not string",
			})
		}
		logrus.Info("Delete file ", iNum)

		mapleFS.RemoveFile(currentDir, file)
		context.JSON(204, nil)
	})

	// rmdir
	r.DELETE("/dirs/:name", func(context *gin.Context) {
		exists, iNum := checkNameUnExists(context)
		if !exists {
			return
		}
		mapleFS.RmDir(mapleFS.IGet(iNum))
		context.JSON(204, nil)
	})

	// create file
	r.POST("/files/:name", func(context *gin.Context) {
		nonExists, name := checkNameExists(context)
		logrus.Info("Create dir ", name)
		if !nonExists {
			return
		}
		iNode := mapleFS.CreateFile(currentDir, []byte(name))
		u := url.URL{}
		u.Host = "127.0.0.1"
		u.Path = "/files/" + name
		context.Header("Location", u.String())

		context.JSON(201, FileSchema{
			name,
			iNode.GetINum(),
			iNode.GetType(),
		})
	})

	// synchronize file
	r.PUT("/files/:name", func(context *gin.Context) {

		exists, iNode := checkNameUnExists(context)
		if !exists {
			return
		}
		if iNode == -1 {
			context.JSON(404, nil)
		}
		newData, exists := context.GetPostForm("text")
		if !exists {

			context.JSON(409, "data not exists")
			return
		}
		logrus.Info("newData:", newData)
		mapleFS.EditFile(mapleFS.IGet(iNode), []byte(newData))
		context.JSON(204, nil)
	})

	// 退出系统
	r.GET("/exit", func(context *gin.Context) {
		currentDir = &rootDir
	})
	r.Run()
}
