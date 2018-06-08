# 文件系统模拟

## Definitions

### Blocks

内存块对应。这里实现了对应的buffer，并且在文件中能够根据 `index`寻找到对应的block区域。这里使用的是寻找对应的buffer。

### Block Groups



### Dictories



### Inodes



### Superblocks



### Symbol links



## Describe

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
	nlink uint8		//

	size uint32		// size of file
	addrs [NDIRECT + 1]uint8	// 直接指向的数据块
	// 多级数据块 -- > 等会儿直接用树组织吧

}
```

