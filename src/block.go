package src

import (
	"log"

	"github.com/sirupsen/logrus"
)

const NBUF = 50

type BufferCache struct {
	Buffers [NBUF]buffer
	// head of buffer, 是一个虚假的表头
	Head buffer
}

var bufferCache BufferCache

func init()  {
	// init bufferCache head
	bufferCache.Head.prev = &bufferCache.Head
	bufferCache.Head.next = &bufferCache.Head

	for _, buf := range bufferCache.Buffers {
		// TODO: make clear how this run
		// 构成链表
		buf.prev = bufferCache.Head.next
		buf.next = &bufferCache.Head
		//buf.sector = -1
		buf.statusFlag = BUF_UNUSED
		bufferCache.Head.next.prev = &buf
		bufferCache.Head.next = &buf
	}


}



func writeToBlockDIO(blockNum uint32, bdata []byte) {
	written, err := fsfd.WriteAt(bdata[:], int64(blockNum * BLOCK_SIZE))
	if err != nil {
		panic(err)
	}
	if written != BLOCK_SIZE {
		panic("Written size error")
	}
}


func readBlockDIO(blockNum uint32) []byte {
	data := make([]byte, BLOCK_SIZE)
	readSize, err := fsfd.ReadAt(data, int64(blockNum * BLOCK_SIZE))
	if err != nil {
		panic(err)
	}
	if readSize != BLOCK_SIZE {
		log.Fatalf("read size is %d and not blocksize", readSize)
	}
	return data
}


// 所有bget的对象都需要已经设置了bitmap
func bget(sector uint16) *buffer {
	head := bufferCache.Head
	var cur *buffer
	for cur = head.next; cur.statusFlag != BUF_UNUSED && cur != &head; cur = cur.next {
		if cur.sector == (sector) {
			return cur
		}
	}
	// 需要申请
	if cur.statusFlag == BUF_UNUSED {
		copy(cur.data[:], readBlockDIO(uint32(sector)))
		cur.sector = sector
		cur.statusFlag = BUF_VALID
		return cur
	}
	// FIFO
	if cur == &head {
		// 淘汰算法
		brelse(head.next)
		ptrBuf := head.prev
		ptrBuf.sector = sector
		ptrBuf.statusFlag = BUF_VALID
		copy(ptrBuf.data[:], readBlockDIO(uint32(sector)))
		cur = ptrBuf
	}
	return cur
}

// 从硬件中申请块
func balloc() *buffer {
	// read bitmap
	lowerB := BBLOCK(0, NINODES)
	upperB := BBLOCK(uint16(BITMAP_BLOCK_NUM), NINODES)
	logrus.Debugf("BITMAP_BLOCK_NUM: %d", BITMAP_BLOCK_NUM)

	for i := lowerB; i <= upperB; i++ {
		bitmap := readBlockDIO(uint32(i))
		var blockmap BlockBitmap
		copy(blockmap[:], bitmap)
		logrus.Println(bitmap)
		for index := 0; index < BLOCK_SIZE; index++ {
			// MD，应该要注意实际的对应关系
			if !blockmap.valid(uint16(index)) {

				// unused
				logrus.Infof("Allocate block at %d sector, pos %d (block %t)", i, index, blockmap.valid(uint16(index)))
				//bitmap[index] = 1
				blockmap.setValid(uint16(index))
				retBuf := buffer{statusFlag:BUF_VALID, sector:uint16(uint16(i-lowerB) * BPB + uint16(index))}
				writeToBlockDIO(uint32(i), blockmap[:])
				return &retBuf
			}
		}
	}
	log.Fatal("Cannot found spare block.")
	return nil
}

func putback(ptrBuf *buffer)  {
	// 丢到队尾
	ptrBuf.next.prev = ptrBuf.prev
	ptrBuf.prev.next = ptrBuf.next
	ptrBuf.next = bufferCache.Head.next
	ptrBuf.prev = &bufferCache.Head

	bufferCache.Head.next.prev = ptrBuf
	bufferCache.Head.next = ptrBuf
}

// 完成读写
func brelse(ptrBuf *buffer)  {
	putback(ptrBuf)

	if ptrBuf.statusFlag == BUF_DIRTY {
		writeToBlockDIO(uint32(ptrBuf.sector), ptrBuf.data[:])
	}

	ptrBuf.statusFlag = BUF_UNUSED
	//ptrBuf.statusFlag &= ~BUF_BUSY
}