package mapleFS

import (
	log "github.com/sirupsen/logrus"
	"bytes"
	"encoding/binary"
	"strings"
	"unsafe"
	"fmt"
)

// 创建目录
func MkDir(name []byte) *inode {
	// TODO: 在本目录下做好查找
	newInode := ialloc()
	log.Info("Alloc INode with inum ", newInode.num)
	newInode.dinodeData.FileType = FILETYPE_DIRECT
	//newInode.dinodeData.Size = 0
	return newInode
}


func MkRootDir() *inode {
	// TODO: 加上对有没有读取节点的检查
	rootDir := MkDir([]byte("root"))

	dirlink(rootDir, []byte("."), rootDir.num)
	dirlink(rootDir, []byte(".."), rootDir.num)
	fsyncINode(rootDir)
	return rootDir
}

// 感觉操作最好还是用缓存 + inode 序号...
func MkDirWithParent(name []byte, parentNodePtr *inode) *inode {
	checkDir(parentNodePtr)
	curNode := MkDir(name)
	dirlink(curNode, []byte(".."), parentNodePtr.num)
	dirlink(curNode, []byte("."), curNode.num)
	dirlink(parentNodePtr, name, curNode.num)
	log.Info("Create INode with number ", curNode.num)
	fsyncINode(curNode)
	fsyncINode(parentNodePtr)
	return curNode
}

func ReadRoot(node *inode) {
	blockBytes := readBlockDIO(IBLOCK(0))
	INODE_SIZE := int(unsafe.Sizeof(Dinode{}))
	var readDi Dinode
	log.Debugf("Read block %d from %d to %d", IBLOCK(0), 0, INODE_SIZE)
	readObject(blockBytes[:INODE_SIZE], &readDi)
	node.dinodeData = readDi
	node.num = 0
	node.ref = 1
	node.valid = 1 	// TODO: make clear what todo is
	if node.dinodeData.FileType != FILETYPE_DIRECT {
		log.Fatalf("FileType Not DIRECT!!!!")
	}
}

func (dirNode *inode) DirIsEmpty() bool {
	checkDir(dirNode)
	// 只有 . .. 对应的条目
	return dirNode.dinodeData.Size == uint32(DIRENT_SIZE) * 2
}

// remove dir, 返回删除是否成功
func RmDir(dirNode *inode) bool {
	checkDir(dirNode)
	if dirNode.DirIsEmpty() {
		unimpletedError()
		//parent := iget(dir)
		return true
	} else {
		return false
	}
}

// 对目录进行链接
func dirlink(dir *inode, destName []byte, inum uint16)  {
	if dir.dinodeData.FileType != FILETYPE_DIRECT {
		log.Fatalf("Type of file error, inode is not dir in dirlink")
	}
	if len(destName) >= DIRSIZ {
		log.Fatalf("Too long file name!")
	}
	var name [DIRSIZ]byte
	copy(name[:], destName)
	dirItem := Dirent{inum, name}
	//log.Debug("Dir size: ", unsafe.Sizeof(dirItem))
	iappend(dir, dirItem)
}

// 对目录取消链接
func dirunlink(dir *inode, destName []byte)  {
	if dir.dinodeData.FileType != FILETYPE_DIRECT {
		log.Fatalf("Type of file error, inode is not dir in dirunlink")
	}
	unimpletedError()
}

func checkDir(node *inode) {
	if node.dinodeData.FileType != FILETYPE_DIRECT {
		log.Fatal("Type is not DIR!")
	}
}

func WalkDir(dir *inode) []*Dirent {
	checkDir(dir)
	var readdir Dirent
	STRUCT_SIZE := int(unsafe.Sizeof(readdir))

	var retArray []*Dirent
	for i := 0; i <= int(dir.dinodeData.Nlink); i++ {
		block := readBlockDIO(dir.dinodeData.Addrs[i])
		log.Debugf("dir inode block %d, read block %d", IBLOCK(uint32(dir.num)), dir.dinodeData.Addrs[i])
		//fmt.Println(block)
		for j := 0;j * STRUCT_SIZE < BLOCK_SIZE && i * BLOCK_SIZE + j * int(STRUCT_SIZE) < int(dir.dinodeData.Size); j++{
			var curDir Dirent

			//log.Debug("From ", j * STRUCT_SIZE, " to ", (j + 1) * STRUCT_SIZE, " data --> ", block[j * STRUCT_SIZE: (j + 1) * STRUCT_SIZE])
			buf := bytes.NewBuffer(block[j * STRUCT_SIZE: (j + 1) * STRUCT_SIZE])
			binary.Read(buf, binary.LittleEndian, &curDir)
			log.Debug("Find dir: ", curDir.String())
			retArray = append(retArray, &curDir)
		}
	}
	return retArray
}

func dirlookup (dir *inode, destName []byte) int {
	// return 0 if not found
	checkDir(dir)
	// TODO: we can optimize it
	s := string(destName)
	//if lSize > DIRSIZ {
	//	return 0
	//}
	for _, d := range WalkDir(dir) {
		// compare
		log.Debug("Comparing ", d.DirName(), " with ", s)
		if strings.Compare(d.DirName(), s) == 0 {
			return int(d.INum)
		} else {
			log.Debug("Comparing ", d.DirName(), " with ", s, " delta: ", strings.Compare(d.DirName(), s))
		}
	}
	return -1
}

// 存储目录项的条目
// TODO: 搞清楚导入导出的机制
type Dirent struct {
	INum uint16
	// 是不是到时候改回rune比较好
	Name [DIRSIZ]byte
}


func (dir *Dirent) DirName() string {
	name := make([]byte, 0)
	var index int
	for index = 0; dir.Name[index] != 0; index++ {
		name = append(name, dir.Name[index])
	}
	return string(name)
}

func (dir *Dirent) String() string {
	return fmt.Sprintf("Dirent(INum: %d, name:%s)", dir.INum, dir.DirName())
}