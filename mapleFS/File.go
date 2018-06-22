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
	return RemoveFile(iNode)
}

func RemoveFile(fileINode *INode) bool {
	if !fileINode.IsFile() {
		// not a file
		log.Infof("INode %d is not a iNode", fileINode.num)
		return false
	}
	for buf := range fileINode.BufferStream() {
		bfree(buf)
	}
	// TODO: impl it
	for index, value := range fileINode.dinodeData.Addrs {
		if value == 0 {
			break
		}
		if index == NDIRECT {
			buf := bget(uint16(value))
			bzero(buf)
			bfree(buf)
		}
		fileINode.dinodeData.Addrs[index] = 0
	}
	fsyncINode(fileINode)
	return false
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
			endvalue = cnt*BLOCK_SIZE - fileINode.dinodeData.Size
		}
		retData = append(retData, buf.data[:endvalue]...)
	}
	return retData
}
