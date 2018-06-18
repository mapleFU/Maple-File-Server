package API

import (
	"github.com/gin-gonic/gin"
	"github.com/mapleFU/TongjiFileLab/mapleFS"
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

	// list files
	r.GET("/files", func(context *gin.Context) {
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
	})

	// mkdir
	r.POST("/dirs/:name", func(context *gin.Context) {

	})

	// rmdir
	r.DELETE("/dirs/:name", func(context *gin.Context) {

	})

	// create file
	r.POST("/files/:name", func(context *gin.Context) {

	})

	// synchronize file
	r.PUT("/files/:name", func(context *gin.Context) {

	})

	// 退出系统
	r.GET("/exit", func(context *gin.Context) {
		currentDir = &rootDir
	})
	r.Run()
}
