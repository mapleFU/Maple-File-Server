# 文件系统模拟

## 实现的接口

* `ls`
* `mkdir`
* `touch`
* `pwd`
* `cd`
* `rm`
* `rmdir`

## 基本的操作和使用



## 架构：

这里才用了自底向上的实现，上一层调用下一层或者别的层次提供的数据结构和接口。以下是关键的名词和接口的层次。

### BasicIO

基础的给程序提供`I/O`的包。

### SuperBlock

```go
type superblock struct {
	Size uint32		// size of blocks
	Nblocks uint32	// number of datablocks
	Ninodes uint32	// number of inodes
}
```

`Superblock` 作为程序的超级块，储存着这一部分的文件系统的基本的信息。

### IDE Driver 层

```
  0: unused
  1: super block (size, ninodes)
  2: log for transactions
  32: array of inodes, packed into blocks
  58: block in-use bitmap (0=free, 1=used)
  59: file/dir content blocks
  end of disk
```



| 区域 |                         功能                         |
| :--: | :--------------------------------------------------: |
|  0   | unused\(本来是启动程序的扇区，但是没有实现对应功能\) |
|  1   |                     super block                      |
|  2   |   log for transactions\(程序对应的log, 暂未实现\)    |
|  x   |                        inodes                        |
|  y   |                 block in use bitmap                  |
|  z   |               file/dir content blocks                |

我们可以对照代码方面的定义：

```go
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
const BPB = BLOCK_SIZE * 8

// BITMAP 占有的 BLOCK 的量
const BITMAP_BLOCK_NUM uint32 = SIZE/(BLOCK_SIZE*8) + 1

// 目录对应的 bytes
const DIRSIZ = 28

const MAX_UINT16 = 65535

type bufferStatus uint8

const (
	BUF_BUSY   bufferStatus = 1 << iota // buffer is locked by some process
	BUF_VALID  bufferStatus = 1 << iota // buffer has been read from disk
	BUF_DIRTY  bufferStatus = 1 << iota // buffer needs to be written to disk
	BUF_UNUSED bufferStatus = 1 << iota
)

const DIRENT_SIZE = uint(unsafe.Sizeof(Dirent{}))

const MAX_UINT32 = 4294967295

const bitblocks uint32 = SIZE/(BLOCK_SIZE*8) + 1
const usedblocks = NINODES/uint32(IPB) + 3 + bitblocks
```

超级块：

```go
type superblock struct {
	size uint32		// size of blocks
	nblocks uint32	// number of datablocks
	ninodes uint32	// number of inodes
}
```

日志系统：暂无

### Block Buffer 层

1. 同步对磁盘的访问
2. 缓存常用的块，提升性能

主要接口：`bread` `bwrite` 从磁盘中写入缓冲区／把缓冲区一块写到磁盘上正确位置。处理完后要用`brelse`

块缓冲利用`LRU`替换

`binit` 从`buf`取出

`binit` 从静态数组`buf` 构建有`NBUF`元素的 `linkedlist` 通过链表访问缓冲区

有 `VALID` `DIRTY` `BUSY` 三种状态

`bget`获得指定扇区缓冲 从磁盘中读取可能会调用`iderw` 扫描缓冲区链表，寻找到对应的，不处于`BUSY` 状态则返回（反之阻塞）

`brelse`  将缓冲区移动到链表头部，清除`B_BUSY` 

block在文件系统还是对应`sector`的一个固定大小的块，这里采用数据结构`buffer`来为访问具体的`block`提供接口和抽象。

```go
/**
disk blocks
  most o/s use blocks of multiple sectors, e.g. 4 KB blocks = 8 sectors
  to reduce book-keeping and seek overheads
  xv6 uses single-sector blocks for simplicity

dev/sector 应该是单一指定的位置？
*/
type buffer struct {
	statusFlag bufferStatus
	// dev, sector 是对应的设备、扇区管理
	dev    uint8  // 设备
	sector uint16 // 扇区 这个程序里面表示所存储的块

	prev, next, qnext *buffer
	// 对应的数据，有着固定的大小
	data [BLOCK_SIZE]byte
}

```

这里`buffer`标定了 `block`的数据块，信息。用`statusFlag`表示这块`buffer`是否读写。`buffer`包抽象了具体的位置，提供了这些抽象的信息。

根据空闲块位图，分配新的块。

`balloc` 分配新的块 `bfree` 释放。 先用`readsb`从磁盘读 `superblock`, `balloc`寻找对应的块，同时清空对应的位。

### INode/Dinode 层

`inode`是 `dinode` 的记录 `ialloc`  申请新的i节点。 `iget` 会遍历 `inode`缓存寻找

`bmap` 会返回对应序号 `inode` 的内容

在操作系统中，dinode` 表示操作系统对文件的标示，一个文件只存在一个`inode`, `inode`用 `inum`这个序号来表示。一下，同时表示又正确的文件的的大小。`mapleFS`中 `dinode` 表示磁盘上的 `inode`, 在文件系统中它的位置是不言自明的。

```go
type Dinode struct {
	FileType uint16 // 文件的类型
	Nlink    uint16 // link 链接的数量

	Major, Minor uint16              // 对应的major minor, 我这里好像没啥用...好吧我他妈把MAJOR当成LINK链接好了,MINOR当成-s link好了
	Size         uint32              // size of file
	Addrs        [NDIRECT + 1]uint32 // 直接指向的数据块，最后一个+1对应的是二级索引

}
```

`inode`是内存中的 `inode`. 保存了`dinode`, 我们的外部操作都是针对inode

```go
/**
inode is dinode in memory

    FS records file info in an "inode" on disk
    FS refers to inode with i-number (internal version of FD)
    inode must have link count (tells us when to free)
    inode must have count of open FDs
    inode deallocation deferred until last link and FD are gone

*/
type INode struct {
	num   uint16     // 对应的序号
	ref   int        // 引用计数
	lock  sync.Mutex // 内容的锁，暂时不会用到
	valid int32      // 是否在disk中被读出

	// copy of disk inode, 指向真实的block信息
	dinodeData Dinode
}
```

### 目录层

目录的 `inode`类型是 `T_DIR`, `dirlookup` `dirlink` `dirunlink` 操作目录

### Dirent

```go
// 存储目录项的条目
// TODO: 搞清楚导入导出的机制
type Dirent struct {
	INum uint16
	// 是不是到时候改回rune比较好
	Name [DIRSIZ]byte
	// 文件的类型
	FileType uint16
}

```

目录项，对应的目录的操作中，会像目录中添加目录项。每个目录项的大小是*32bytes*, 每个block能存储固定的dirent作为目录的记录。

### 文件描述符／系统调用层

`sys_link` `sys_unlink` `nameiparent` `dirlookup`

### File

```go
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

```

*FIle*, 作为系统的文件，同样用`inode`表示，可以往里面同步内容，添加信息。