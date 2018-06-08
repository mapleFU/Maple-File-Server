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

import "unsafe"

/**
// File system super block
struct superblock {
	uint size;         // Size of file system image (blocks)
	uint nblocks;      // Number of data blocks
	uint ninodes;      // Number of inodes.
};
 */

 // 操作系统块的常量
const ROOT_INODE_NUM = 1
const BLOCK_SIZE = 512

type superblock struct {
	size uint32		// size of blocks
	nblocks uint32	// number of datablocks
	ninodes uint32	// number of inodes
}

func readsb(superblock *superblock) {
	panic("not implemented.")
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
	FREE = iota
	FILE
	DIRECT

)

type dinode struct {
	fileType uint16	// 文件的类型
	nlink uint16		// link 链接的数量

	major, minor uint16	// 对应的major minor, 我这里好像没啥用
	size uint32		// size of file
	addrs [NDIRECT + 1]uint32	// 直接指向的数据块，最后一个+1对应的是二级索引

}

// TODO:弄出一个实际的块...? 这个安全吗？
// 一个BLOCK能存储的INODE的数目
const IPB = BLOCK_SIZE / unsafe.Sizeof(dinode{})

// Block containing inode i
// 给出 index, 描述出index block对应的位置，SUPERBLOCK == 1
func IBLOCK(i uint16) uint16 {
	return i / uint16(IPB) + 2
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
 
type dirent struct {
	inum uint16
	// 是不是到时候改回rune比较好
	name [DIRSIZ]byte
}

const DIRENT_SIZE = unsafe.Sizeof(dirent{})

// 对目录进行链接
func dirlink(dir *dirent, bytes []byte, inode *inode)  {
	unimpletedError()
}

// 对目录取消链接
func dirunlink(dirent *dirent, name []byte, pinode *inode)  {
	unimpletedError()
}

func bmap(inode *inode, bn uint32) uint32 {
	unimpletedError()
	return uint32(1)
}