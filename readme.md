# 文件系统模拟

## Requirements

`create` `open` `close` `write` `read` `unlink` 

`cd` `pwd` `mkdir` `rmdir` 

## Definitions

### Blocks

内存块对应。这里实现了对应的buffer，并且在文件中能够根据 `index`寻找到对应的block区域。这里使用的是寻找对应的buffer。

### Block Groups



### Dictories



### Inodes



### Superblocks



### Symbol links



## Describe

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



| 区域 | 功能 |
| :---: | :--: |
| 0 | unused |
| 1 | super block |
| 2 | log for transactions |
| x | inodes |
| y | block in use bitmap |
| z | file/dir content blocks |

我们可以对照代码方面的定义：



```go
const ROOT_INODE_NUM = 1
const BLOCK_SIZE = 512
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

`inodes`:

系统的`inode`

```go
const (
	FREE = iota
	FILE
	DIRECT
)

type dinode struct {
	fileType uint8	// 文件的类型
	nlink uint8		// link 链接的数量

	major, minor uint8	// 对应的major minor, 我这里好像没啥用
	size uint32		// size of file
	addrs [NDIRECT + 1]uint8	// 直接指向的数据块
	// 多级数据块 -- > 等会儿直接用树组织吧
}

// 一个BLOCK能存储的INODE的数目
const IPB = BLOCK_SIZE / unsafe.Sizeof(dinode{})

// 给出 index, 描述出index block对应的位置
func IBLOCK(i uint8) uint8 {
	return i / uint8(IPB) + 2
}

// Bitmap bits per block, 每块需要对应一个长度为 BLOCK_SIZE * uint8 的字段
const BPB  = BLOCK_SIZE * 8

// 这个应该表示的是bitmap block 对应的位置, B表示的是第几个块, 对应的是哪个位置
func BBLOCK(b, ninodes uint8) uint8 {
	// 本来应该是 + 2, 但是实际上这里至少有一个block会被INODES TABLE占用，所以 + 3
	return uint8(b / BPB + ninodes / uint8(IPB) + 3)
}

// 目录信息
const DIRSIZ = 14
 
type dirent struct {
	inum uint8
	// 是不是到时候改回rune比较好
	name [DIRSIZ]byte
}
```



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



### 数据结构

同时对于内存中的`inode` 有这样的描述

```go
type inode struct {
	num uint32	// 对应的序号
	ref int	// 引用计数
	lock sync.Mutex	// 内容的锁，暂时不会用到
	valid int32	// 是否在disk中被读出

	// copy of disk inode, 指向真实的block信息
	dinodeData dinode
}
```



### 块分配层

根据空闲块位图，分配新的块。

`balloc` 分配新的块 `bfree` 释放。 先用`readsb`从磁盘读 `superblock`, `balloc`寻找对应的块，同时清空对应的位。

### inode 层

`inode`是 `dinode` 的记录 `ialloc`  申请新的i节点。 `iget` 会遍历 `inode`缓存寻找

`bmap` 会返回对应序号 `inode` 的内容

### 目录层

目录的 `inode`类型是 `T_DIR`, `dirlookup` `dirlink` `dirunlink` 操作目录

### 文件描述符／系统调用层

`sys_link` `sys_unlink` `nameiparent` `dirlookup`