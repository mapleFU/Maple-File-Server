package src

type bufferStatus uint8

const (
	BUF_BUSY bufferStatus = 1 << iota 	// buffer is locked by some process
	BUF_VALID bufferStatus = 1 << iota	// buffer has been read from disk
	BUF_DIRTY bufferStatus = 1 << iota	// buffer needs to be written to disk
)

const BUFFER_SIZE = 1024

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
	sector uint8	// 扇区？

	prev, next, qnext *buffer
	// 对应的数据，有着固定的大小
	data [BUFFER_SIZE]byte
} 
