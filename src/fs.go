/**
// On-disk file system format.
// Both the kernel and user programs use this header file.

// Block 0 is unused.
// Block 1 is super block.
// Inodes start at block 2.

#define ROOTINO 1  // root i-number
#define BSIZE 512  // block size
*/

package src

import (
	"unsafe"
	"log"
	"encoding/binary"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
)

/**
// File system super block
struct superblock {
	uint size;         // Size of file system image (blocks)
	uint nblocks;      // Number of data blocks
	uint ninodes;      // Number of inodes.
};
 */

 // 操作系统块的常量
const ROOT_INODE_NUM uint32 = 1
const BLOCK_SIZE = 512

type superblock struct {
	Size uint32		// size of blocks
	Nblocks uint32	// number of datablocks
	Ninodes uint32	// number of inodes
}

// 初始化传入的 SUPER BLOCK 指针
func readsb(unInitSptr *superblock) {

	datas := make([]byte, BLOCK_SIZE)
	readSize, err := fsfd.ReadAt(datas, BLOCK_SIZE * 1)
	if readSize != BLOCK_SIZE || err != nil {
		log.Fatalf("Only read %d\n", readSize)
	}
	buf := bytes.NewBuffer(datas[:unsafe.Sizeof(superblock{})])
	err = binary.Read(buf, binary.LittleEndian, unInitSptr)
	if err != nil {
		panic(err)
	}
}
/**
xv6 blocks:
// Inodes per block.
#define IPB           (BSIZE / sizeof(struct dinode))

// Block containing inode i
#define IBLOCK(i)     ((i) / IPB + 2)

// Bitmap bits per block
#define BPB           (BSIZE*8)

// Block containing bit for block b
#define BBLOCK(b, ninodes) (b/BPB + (ninodes)/IPB + 3)
 */

/**
xv6:
// On-disk inode structure
struct dinode
	short type;           // File type
	short major;          // Major device number (T_DEV only)
	short minor;          // Minor device number (T_DEV only)
	short nlink;          // Number of links to inode in file system
	uint size;            // Size of file (bytes)
	uint addrs[NDIRECT+1];   // Data block addresses
};

direct and indirect blocks

example:
  how to find file's byte 8000?
  logical block 15 = 8000 / 512
  3rd entry in the indirect block

each i-node has an i-number
  easy to turn i-number into inode
  inode is 64 bytes long
  byte address on disk: 32*512 + 64*inum
 */

// 直接的指针
const (
	// 直接指针的数目
	NDIRECT = 12
	// 多级索引的最大上限。因为这里只有一个块作为二级索引
	NINDIRECT = BLOCK_SIZE / unsafe.Sizeof(uint32(0))
	// 最多的文件指针？
	MAX_FILE = NDIRECT + NINDIRECT
)

const (
	FILETYPE_FREE = iota
	FILETYPE_FILE
	FILETYPE_DIRECT

)

type Dinode struct {
	FileType uint16	// 文件的类型
	Nlink uint16		// link 链接的数量

	Major, Minor uint16	// 对应的major minor, 我这里好像没啥用...好吧我他妈把MAJOR当成LINK链接好了,MINOR当成-s link好了
	Size uint32		// size of file
	Addrs [NDIRECT + 1]uint32	// 直接指向的数据块，最后一个+1对应的是二级索引

}

// TODO:弄出一个实际的块...? 这个安全吗？
// 一个BLOCK能存储的INODE的数目
const IPB = BLOCK_SIZE / unsafe.Sizeof(Dinode{})

// Block containing inode i
// 给出 index, 描述出index block对应的位置，SUPERBLOCK == 1
func IBLOCK(i uint32) uint32 {
	return i / uint32(IPB) + 2
}

// Bitmap bits per block
const BPB  = BLOCK_SIZE * 8

/**
Block containing bit for block b
ninodes means ninode index
b means bios (in )
 */
 // 这个应该表示的是bitmap block 对应的位置, B表示的是第几个块, 对应的是哪个位置
func BBLOCK(b uint16, ninodes uint8) uint8 {

	// 本来应该是 + 2, 但是实际上这里至少有一个block会被INODES TABLE占用，所以 + 3
	return uint8(uint8(b / BPB) + ninodes / uint8(IPB) + 3)
}


/**
XV6 目录
// Directory is a file containing a sequence of dirent structures.
#define DIRSIZ 14

struct dirent {
	ushort inum;
	char name[DIRSIZ];
};
 */

const DIRSIZ = 14
const MAX_UINT16 = 65535	// 表示删除的记录

// 存储目录项的条目
type dirent struct {
	inum uint16
	// 是不是到时候改回rune比较好
	name [DIRSIZ]byte
}

const DIRENT_SIZE = unsafe.Sizeof(dirent{})

// 创建目录
func mkdir(name []byte) *inode {
	// TODO: 在本目录下做好查找
	newInode := ialloc()
	newInode.dinodeData.FileType = FILETYPE_DIRECT

	return newInode
}

// remove dir
func rmdir(name []byte) *dirent {
	unimpletedError()
	return nil
}

// search for dir
func search(fileURI []byte)  {
	
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
	iappend(dir, dirent{inum, name})
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

func walkdir(dir *inode) {
	checkDir(dir)
	var readdir dirent
	STRUCT_SIZE := int(unsafe.Sizeof(readdir))
	for i := 0; i < int(dir.dinodeData.Nlink); i++ {
		block := readBlockDIO(dir.dinodeData.Addrs[i])
		logrus.Debugf("read block %d", dir.dinodeData.Addrs[i])

		for j := 0;j * STRUCT_SIZE < BLOCK_SIZE && i * BLOCK_SIZE + j * int(STRUCT_SIZE) < int(dir.dinodeData.Size); j++{

			buf := bytes.NewBuffer(block[j * STRUCT_SIZE: (j + 1) * STRUCT_SIZE])
			binary.Read(buf, binary.LittleEndian, &readdir)
			fmt.Println(readdir)
		}
	}
}

func dirlookup (dir *inode, destName []byte) {
	checkDir(dir)

}

func bmap(inode *inode, bn uint32) uint32 {
	unimpletedError()
	return uint32(1)
}

