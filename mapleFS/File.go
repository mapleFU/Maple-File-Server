package mapleFS

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
	iNode := ialloc()
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
	unimpletedError()
}
