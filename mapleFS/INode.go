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

func (node *INode) GetINum() uint16 {
	return node.num
}

func (node *INode) GetType() string {
	switch node.dinodeData.FileType {
	case FILETYPE_DIRECT:
		return "DIRECT"
	case FILETYPE_FILE:
		return "FILE"
	case FILETYPE_FREE:
		return "UNKNOWN"
	default:
		panic("Type of dir is unexcepted.")
	}
}

func IAlloc() *INode {
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
func IGet(inodeIndex int) *INode {
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

func IAddblock(node *INode) {
	if node.dinodeData.Nlink < NDIRECT {
		blockBuf := balloc()
		node.dinodeData.Addrs[node.dinodeData.Nlink] = uint32(blockBuf.sector)
		log.Infof("Create block with sector %d", blockBuf.sector)
	} else {
		unimpletedError()
	}
}

// 向inode中插入数据
func IAppend(node *INode, dataStruct interface{}) {
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
			IAddblock(node)
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
			IAppend(node, datas[BLOCK_SIZE-bios:])
		}
		// 写回
		writeToBlockDIO(linkAddr, blockData)
	} else {
		// TODO: 实现间接索引
		// 间接 暂时没有实现
		unimpletedError()
	}

}

// 我总觉得这个函数会出事
// 全部修改一个节点的信息
func IModify(node *INode, newData []byte) {
	var editedBytes uint32 = 0 // 编辑过的byte
	var remainBytes = uint32(len(newData))
	var ifEnd = false // 读写是否停止
	var cnt = 0
	// 先用掉已经申请的块
	for buf := range node.BufferStream() {
		if !ifEnd {
			cnt++
			// 先编辑存在的buf
			if remainBytes > BLOCK_SIZE {
				copy(buf.data[:], newData[editedBytes:editedBytes+BLOCK_SIZE])
				editedBytes += BLOCK_SIZE
				remainBytes -= BLOCK_SIZE
			} else {
				bzero(buf)
				copy(buf.data[:remainBytes], newData[editedBytes:editedBytes+remainBytes])
				editedBytes += remainBytes
				remainBytes = 0
				// release node data
				ifEnd = false

			}
		} else {
			bfree(buf)
		}
	}
	oldSize := node.dinodeData.Size
	// 如果没有完成，iappend 会妥善修改内容
	node.dinodeData.Size = editedBytes
	if !ifEnd {
		// 没有结束，添加内容
		for remainBytes > 0 {
			IAppend(node, newData[editedBytes:])
		}
	} else {
		// 内容过剩，处理NLink
		currentUsed := editedBytes / BLOCK_SIZE
		if editedBytes%BLOCK_SIZE != 0 {
			currentUsed++
		}
		oldUsed := oldSize / BLOCK_SIZE
		if oldSize%BLOCK_SIZE != 0 {
			oldSize++
		}
		// 使用
		node.dinodeData.Nlink = uint16(currentUsed)
		if currentUsed >= NDIRECT {
			var secondUsed = int(currentUsed - NDIRECT)
			buf := bget(uint16(node.dinodeData.Addrs[NDIRECT]))
			intSize := int(unsafe.Sizeof(uint32(0)))
			var secnt int
			var nodeNum uint32
			for secnt < int(oldUsed-currentUsed) {
				readObject(buf.data[intSize*(secondUsed+secnt):intSize*(secondUsed+secnt+1)], &nodeNum)
				// 释放多余的块
				bfree(bget(uint16(nodeNum)))
				secnt++
			}
			copy(buf.data[intSize*secondUsed:], zeroBuf[intSize*secondUsed:])
		} else {
			// 初始化 ADDRS 字段
			node.dinodeData.Nlink = uint16(currentUsed)
			for i := range node.dinodeData.Addrs {
				if i > int(node.dinodeData.Nlink) {
					node.dinodeData.Addrs[i] = 0
				}
			}
		}
	}
	fsyncINode(node)
}

func (node *INode) BufferStream() <-chan *buffer {
	bufChan := make(chan *buffer)
	// 向 chan 发送信息
	go func() {

		var index uint16
		for index = 0; index <= node.dinodeData.Nlink && index < NDIRECT; index++ {
			log.Info("Send buf no:", index)
			bufChan <- bget(uint16(node.dinodeData.Addrs[int(index)]))
		}
		if node.dinodeData.Nlink == NDIRECT {
			// 读取
			secondBuf := bget(uint16(node.dinodeData.Addrs[NDIRECT]))
			// 里面的元素数目
			nodeSize := int(unsafe.Sizeof(uint32(0)))
			// 不可能等于0，所以一个个读
			var readSecondIndex uint32 // 读出来的索引地址
			var secIndex int           // 次级对应的index
			for secIndex < BLOCK_SIZE/nodeSize {
				readObject(secondBuf.data[secIndex*nodeSize:(secIndex+1)*nodeSize], &readSecondIndex)
				if readSecondIndex == 0 {
					close(bufChan)
					return
				} else {
					bufChan <- bget(uint16(readSecondIndex))
				}
			}
		}
		close(bufChan)
	}()
	return bufChan
}
