package API

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mapleFU/TongjiFileLab/mapleFS"
	"log"
	"net/url"
)

type FileSchema struct {
	FileName string `json:"file_name"`
	INum     uint16 `json:"i_num"`
	FileType string `json:"file_type"`
}

func main() {
	r := gin.Default()
	var rootDir mapleFS.INode
	mapleFS.ReadRoot(&rootDir)
	currentDir := &rootDir

	// 检查名称是否存在，如果存在折返回 false
	checkNameExists := func(context *gin.Context) (bool, string) {
		name, exists := context.GetPostForm("name")
		if !exists {
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
		name, exists := context.GetPostForm("name")
		if !exists {
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

		context.JSON(200, fileDatas)
	})

	// read concrete file, like cat
	r.GET("/files/:name", func(context *gin.Context) {
		// watch a file
		fileName := context.Param("name")
		iNum := mapleFS.Dirlookup(currentDir, []byte(fileName))
		if iNum == -1 {
			context.String(404, "Resource not found.")
		}
		var resultMap = map[string]string{
			"text": string(mapleFS.ReadFileFromINum(uint16(iNum))),
		}

		context.JSON(200, resultMap)
	})

	// mkdir
	r.POST("/dirs/:name", func(context *gin.Context) {
		exists, name := checkNameExists(context)
		if exists {
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

	})
	// rmdir
	r.DELETE("/dirs/:name", func(context *gin.Context) {
		exists, iNum := checkNameUnExists(context)
		if !exists {
			return
		}
		mapleFS.RmDir(mapleFS.IGet(iNum))
		context.JSON(404, nil)
	})

	// create file
	r.POST("/files/:name", func(context *gin.Context) {
		exists, name := checkNameExists(context)
		if exists {
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
		exists, name := checkNameExists(context)
		if !exists {
			return
		}
		iNode := mapleFS.Dirlookup(currentDir, []byte(name))
		if iNode == -1 {
			context.JSON(404, nil)
		}
		newData, exists := context.GetPostForm("data")
		if !exists {
			context.JSON(409, "data not exists")
		}

		mapleFS.EditFile(mapleFS.IGet(iNode), []byte(newData))
		context.JSON(204, nil)
	})

	// 退出系统
	r.GET("/exit", func(context *gin.Context) {
		currentDir = &rootDir
	})
	r.Run()
}
