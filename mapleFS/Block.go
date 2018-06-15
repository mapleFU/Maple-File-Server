package mapleFS

import (
	"github.com/sirupsen/logrus"
	"log"
)

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
	dev uint8		// 设备
	sector uint16	// 扇区 这个程序里面表示所存储的块

	prev, next, qnext *buffer
	// 对应的数据，有着固定的大小
	data [BLOCK_SIZE]byte
}

// 所有bget的对象都需要已经设置了bitmap
// 可以考虑加上cache.
// 默认get的对象都alloc过
func bget(sector uint16) *buffer {
	var cur buffer
	copy(cur.data[:], readBlockDIO(uint32(sector)))
	cur.sector = sector
	cur.statusFlag = BUF_VALID
	return &cur
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


// 完成读写
func brelse(ptrBuf *buffer)  {
	if ptrBuf.statusFlag == BUF_DIRTY {
		writeToBlockDIO(uint32(ptrBuf.sector), ptrBuf.data[:])
	}

	ptrBuf.statusFlag = BUF_UNUSED
	//ptrBuf.statusFlag &= ~BUF_BUSY
}
