package mapleFS

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"sync"
	"unsafe"
)

type Dinode struct {
	FileType uint16 // 文件的类型
	Nlink    uint16 // link 链接的数量

	Major, Minor uint16              // 对应的major minor, 我这里好像没啥用...好吧我他妈把MAJOR当成LINK链接好了,MINOR当成-s link好了
	Size         uint32              // size of file
	Addrs        [NDIRECT + 1]uint32 // 直接指向的数据块，最后一个+1对应的是二级索引

}

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

func ialloc() *INode {
	var inodeBlocks []byte
	var sb superblock
	// read super block
	readsb(&sb)
	lowerB := IBLOCK(0)
	upperB := IBLOCK(NINODES)
	INODE_LENGTH := int(unsafe.Sizeof(Dinode{}))

	for i := lowerB; i <= upperB; i++ {
		// 读取对应的block
		inodeBlocks = readBlockDIO(i)
		for innerInum := 0; innerInum < int(IPB); innerInum++ {
			// read dinode
			var readDi Dinode
			curBase := innerInum * INODE_LENGTH // 目前这一个的基址
			readObject(inodeBlocks[curBase:(innerInum+1)*INODE_LENGTH], &readDi)
			//log.Info(readDi)
			// not unequal
			if readDi.Size == MAX_UINT32 {
				// TODO: impletes this
				readDi.Major++
				readDi.Size = 0
				curNode := INode{num: uint16((i-lowerB)*uint32(IPB) + uint32(innerInum)), ref: 1, dinodeData: readDi}

				// 写回数据
				buf := bytes.Buffer{}
				err := binary.Write(&buf, binary.LittleEndian, curNode.dinodeData)
				if err != nil {
					panic(err)
				}

				log.Info("Debug: size ", curNode.dinodeData.Size)
				copy(inodeBlocks[curBase:(innerInum+1)*INODE_LENGTH], buf.Bytes())
				writeToBlockDIO(i, inodeBlocks)

				//icachemap[int(i * uint32(IPB)) + innerInum] = &curNode

				// 增加inode
				innerInum++
				//fsyncINode(&curNode)
				return &curNode
			}
		}
	}
	log.Fatalf("Not found data!")
	return nil
}

// TODO: fill this one to complete the demo
func (dinode *Dinode) toINode() *INode {
	var retNode INode
	retNode.dinodeData = *dinode

	return &retNode
}

// 遍历缓存找到对应的项
func iget(inodeIndex int) *INode {
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
	thisINode.num = uint16(inodeIndex)
	//icachemap[inodeIndex] = thisINode
	return thisINode
}

func iaddblock(node *INode) {
	if node.dinodeData.Nlink < NDIRECT {
		blockBuf := balloc()
		node.dinodeData.Addrs[node.dinodeData.Nlink] = uint32(blockBuf.sector)
		log.Infof("Create block with sector %d", blockBuf.sector)
	} else {
		unimpletedError()
	}
}

// 向inode中插入数据
func iappend(node *INode, dataStruct interface{}) {
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
		//log.Info("Write ", len(datas), " of data to INode ", node.num, " begin at ", bios, "Data: ", datas)
		if int(bios)+len(datas) < BLOCK_SIZE {
			copy(blockData[bios:int(bios)+len(datas)], datas)
			node.dinodeData.Size += uint32(len(datas))
		} else if int(bios)+len(datas) == BLOCK_SIZE {
			copy(blockData[bios:int(bios)+len(datas)], datas)
			node.dinodeData.Size += uint32(len(datas))
			node.dinodeData.Nlink++ // 添加指针计数
		} else {
			// 部分拷贝
			copy(blockData[bios:BLOCK_SIZE], datas[0:BLOCK_SIZE-bios])
			node.dinodeData.Size += uint32(BLOCK_SIZE - bios)
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

// 全部修改一个节点的信息
func imodify(node *INode, newData []byte) {
	unimpletedError()
}
