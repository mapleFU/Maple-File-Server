package src

import (
	"sync"
	"unsafe"
	"encoding/binary"
	"bytes"
	"log"

	"github.com/sirupsen/logrus"
)

/**
inode is dinode in memory

    FS records file info in an "inode" on disk
    FS refers to inode with i-number (internal version of FD)
    inode must have link count (tells us when to free)
    inode must have count of open FDs
    inode deallocation deferred until last link and FD are gone

 */
type inode struct {
	num uint16	// 对应的序号
	ref int	// 引用计数
	lock sync.Mutex	// 内容的锁，暂时不会用到
	valid int32	// 是否在disk中被读出

	// copy of disk inode, 指向真实的block信息
	dinodeData Dinode
}

var icachemap [NINODES]*inode

func init() {
	// 初始化 NODE 的 cache
}

// 析构函数
func (inode *inode) destruct()  {

}

// the system call to create inode
/**
遍历磁盘上的结构，寻找到空闲的结构，标注并返回

/**
	pos, err := fsfd.Seek(BLOCK_SIZE * 1, 0)
	if err != nil {
		panic(err)
	}
	if pos != BLOCK_SIZE * 1 {
		log.Fatalf("Move to %d in readsb", pos)
	}
	datas := make([]byte, BLOCK_SIZE)
	readSize, err := fsfd.Read(datas)
	if readSize != BLOCK_SIZE || err != nil {
		log.Fatalf("Only read %d\n", readSize)
	}
	buf := bytes.NewBuffer(datas[:unsafe.Sizeof(superblock{})])
	err = binary.Read(buf, binary.LittleEndian, unInitSptr)
	if err != nil {
		panic(err)
	}


 */

 var inodeNum = ROOT_INODE_NUM

func ialloc() *inode {
	var inodeBlocks []byte
	var sb superblock
	// read super block
	readsb(&sb)
	lowerB := IBLOCK(0)
	upperB := IBLOCK(sb.Ninodes - 1)
	INODE_LENGTH := int(unsafe.Sizeof(Dinode{}))
	// 读取的dinode
	var readDi Dinode
	for i := lowerB; i <= upperB; i++ {
		// 读取对应的block
		inodeBlocks = readBlockDIO(i)
		for innerInum := 0; innerInum < int(IPB); innerInum++  {
			// read dinode
			curBase := innerInum * INODE_LENGTH		// 目前这一个的基址
			didata := inodeBlocks[curBase: (innerInum+1) * INODE_LENGTH]
			// parse didata
			buf := bytes.NewBuffer(didata)
			if len(didata) != INODE_LENGTH {
				panic("Size error!")
			}
			log.Printf("Read %d", i)
			err := binary.Read(buf, binary.LittleEndian, &readDi)
			if err != nil {
				panic(err)
			}
			// not unequal
			if readDi.Size == MAX_UINT32 {
				// TODO: impletes this
				curNode := inode{num:uint16(inodeNum), ref:1, dinodeData:readDi}
				// 增加计数, 改变占用
				readDi.Major++
				readDi.Size = 0
				// 写回数据
				buf := bytes.Buffer{}
				err := binary.Write(&buf, binary.LittleEndian, readDi)
				if err != nil {
					panic(err)
				}
				wData := buf.Bytes()
				for index := 0; index < INODE_LENGTH; index++ {
					inodeBlocks[curBase+index] = wData[index]
				}
				writeToBlockDIO(i, inodeBlocks)

				icachemap[int(i * uint32(IPB)) + innerInum] = &curNode

				// 增加inode
				innerInum++
				return &curNode
			}
		}
	}
	log.Fatalf("Not found data!")
	return nil
}

// TODO: fill this one to complete the demo
func (dinode *Dinode) toINode() *inode {
	var retNode inode
	retNode.dinodeData = *dinode

	return &retNode
}

// 遍历缓存找到对应的项
func iget( inodeIndex int) *inode {
	if icachemap[inodeIndex] != nil {
		return icachemap[inodeIndex]
	}
	// TODO: can we abstract this?
	// 读取文件，数据同步
	imap := readBlockDIO(IBLOCK(uint32(inodeIndex)))
	privateIndex := inodeIndex % int(IPB)
	begPos := privateIndex * int(unsafe.Sizeof(Dinode{}))
	endPos := begPos + int(unsafe.Sizeof(Dinode{}))
	var dinode Dinode

	err := binary.Read(bytes.NewBuffer(imap[begPos:endPos]), binary.LittleEndian, &dinode)
	if err != nil {
		panic(err)
	}
	thisINode := dinode.toINode()

	return thisINode
}

// 向文件中写入 inode
func fsyncINode(node *inode) {
	inodeIndex := int(IBLOCK(uint32(node.num)))

	imap := readBlockDIO(IBLOCK(uint32(inodeIndex)))
	privateIndex := inodeIndex % int(IPB)
	begPos := privateIndex * int(unsafe.Sizeof(Dinode{}))
	endPos := begPos + int(unsafe.Sizeof(Dinode{}))
	//var dinode Dinode
	buf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(buf, binary.LittleEndian, node.dinodeData)
	if err != nil {
		panic(err)
	}
	copy(imap[begPos:endPos], buf.Bytes())
	//thisINode := dinode.toINode()
}

func iaddblock(node *inode) {
	if node.dinodeData.Nlink < NDIRECT {
		blockBuf := balloc()
		node.dinodeData.Addrs[node.dinodeData.Nlink] = uint32(blockBuf.sector)
		logrus.Infof("Create block with sector %d", blockBuf.sector)
	} else {
		unimpletedError()
	}
}

// 向inode中插入数据
func iappend(node *inode, dataStruct interface{})  {
	//var newIndex uint16
	var datas, byteData []byte
	byteData, ok := dataStruct.([]byte)
	if ok {
		datas = byteData
	} else {
		// TODO: make clear how buf run
		buf := bytes.NewBuffer(byteData)
		binary.Write(buf, binary.LittleEndian, dataStruct)
		// 这个合理么
		datas = buf.Bytes()
	}

	// test
	// 把 Data 写入blocks
	if node.dinodeData.Nlink < NDIRECT {
		linkAddr := node.dinodeData.Addrs[node.dinodeData.Nlink]
		if linkAddr == 0 {
			// 需要申请空间
			iaddblock(node)
			linkAddr = node.dinodeData.Addrs[node.dinodeData.Nlink]
		}

		blockData := readBlockDIO(linkAddr)
		bios := node.dinodeData.Size % BLOCK_SIZE
		logrus.Info("Write ", len(datas), " of data to INode ", node.num, " begin at ", bios, "Data: ", datas)
		if int(bios) + len(datas) < BLOCK_SIZE {
			copy(blockData[bios:int(bios) + len(datas)], datas)
			node.dinodeData.Size += uint32(len(datas))
		} else if int(bios) + len(datas) == BLOCK_SIZE{
			copy(blockData[bios:int(bios) + len(datas)], datas)
			node.dinodeData.Size += uint32(len(datas))
			node.dinodeData.Nlink++	// 添加指针计数
		} else {
			// 部分拷贝
			copy(blockData[bios:BLOCK_SIZE], datas[0: BLOCK_SIZE-bios])
			node.dinodeData.Size += uint32(BLOCK_SIZE-bios)
			node.dinodeData.Nlink++
			// 继续 append, 调用别的部分
			iappend(node, datas[BLOCK_SIZE-bios:])
		}
		// 写回
		writeToBlockDIO(linkAddr, blockData)
	} else {
		// TODO: 实现间接索引
		// 间接 暂时没有实现
		unimpletedError()
	}

}