package mapleFS

import log "github.com/sirupsen/logrus"

type FileType uint

// https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
const (
	FD_NONE  = iota
	FD_PIPE  = iota
	FD_INODE = iota
)

type FsFile struct {
	ref      int // ref cnt
	inodePtr *INode

	readable  bool
	writeable bool
}

// 创建文件
func CreateFile(parentDir *INode, fileName []byte) *INode {
	iNode := IAlloc()
	iNode.dinodeData.FileType = FILETYPE_FILE
	iNode.dinodeData.Size = 0
	fsyncINode(iNode)

	// link dir
	dirlink(parentDir, fileName, iNode.num, iNode.dinodeData.FileType)
	fsyncINode(parentDir)
	return iNode
}

// 判断是否是文件
func (node *INode) IsFile() bool {
	return node.dinodeData.FileType == FILETYPE_FILE
}

func AppendFile(fileINode *INode, newData []byte) {
	unimpletedError()
}

func EditFile(fileINode *INode, newData []byte) {
	log.Info("Edit with ", len(newData), " bytes data.")
	if !fileINode.IsFile() {
		log.Fatalf("File iNode is not file in EditFile")
	}
	IModify(fileINode, newData)
}

func RemoveFileWithName(parentNode *INode, fileName []byte, newData []byte) bool {

	inodeNum := Dirlookup(parentNode, fileName)
	if inodeNum == -1 {
		return false
	}
	iNode := IGet(inodeNum)
	return RemoveFile(parentNode, iNode)
}

func RemoveFile(currentDir *INode, fileINode *INode) bool {
	dirunlink(currentDir, fileINode.num)
	IFree(fileINode)
	return true
}

func ReadFileFromINum(iNum uint16) []byte {
	fINode := IGet(int(iNum))
	if fINode == nil {
		return nil
	} else {
		return ReadFile(fINode)
	}
}

func ReadFile(fileINode *INode) []byte {
	if !fileINode.IsFile() {
		log.Fatal("INode ", fileINode, " is not file.")
	}

	var retData []byte
	// read non second data
	var cnt uint32 = 0
	var endvalue uint32 = BLOCK_SIZE
	for buf := range fileINode.BufferStream() {

		cnt++
		if cnt*BLOCK_SIZE > fileINode.dinodeData.Size {
			endvalue = fileINode.dinodeData.Size - (cnt-1)*BLOCK_SIZE
		}
		retData = append(retData, buf.data[:endvalue]...)
		log.Info("Read ", endvalue, " bytes.")
	}
	return retData
}
