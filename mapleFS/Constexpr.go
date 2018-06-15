package mapleFS

import "unsafe"

const FS_IMG_FILE = "maple-xv6.dmg"

const (
	// 总共大小
	SIZE = 1024
	// INODES 的数量
	NINODES = 200
	// BLOCKS 的数量
	NBLOCKS = 995
)

const ROOT_INODE_NUM uint32 = 0
const BLOCK_SIZE = 512

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

// TODO:弄出一个实际的块...? 这个安全吗？
// 一个BLOCK能存储的INODE的数目
const IPB = BLOCK_SIZE / unsafe.Sizeof(Dinode{})

// Bitmap bits per block
const BPB  = BLOCK_SIZE * 8

// BITMAP 占有的 BLOCK 的量
const BITMAP_BLOCK_NUM uint32 = SIZE / (BLOCK_SIZE * 8) + 1

// 目录对应的 bytes
const DIRSIZ = 14

const MAX_UINT16 = 65535

type bufferStatus uint8

const (
	BUF_BUSY bufferStatus = 1 << iota 	// buffer is locked by some process
	BUF_VALID bufferStatus = 1 << iota    // buffer has been read from disk
	BUF_DIRTY bufferStatus = 1 << iota 	// buffer needs to be written to disk
	BUF_UNUSED bufferStatus = 1 << iota
)

const DIRENT_SIZE = uint(unsafe.Sizeof(Dirent{}))

const MAX_UINT32  = 4294967295

const bitblocks uint32 = SIZE / (BLOCK_SIZE * 8) + 1
const usedblocks = NINODES / uint32(IPB) + 3 + bitblocks