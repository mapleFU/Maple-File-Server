package mapleFS

import (
	bytes "bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"os"
	"unsafe"
)

func xuint16(x uint16) uint16 {
	bs := make([]byte, 16)

	bs[0] = byte(x)
	bs[1] = byte(x >> 8)

	return binary.LittleEndian.Uint16(bs)
}

func xuint32(x uint32) uint32 {
	bs := make([]byte, 32)

	bs[0] = byte(x)
	bs[1] = byte(x >> 8)
	bs[2] = byte(x >> 16)
	bs[3] = byte(x >> 24)

	return binary.LittleEndian.Uint32(bs)
}

// 真实指向文件的指针
var fsfd *os.File

// 初始化，为创建／初始化文件系统打开镜像文件，设置日志
func InitMkfs() {
	var err error
	// 基本的信息
	fsfd, err = os.OpenFile(FS_IMG_FILE, os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalln("cannot open the file")
	}
	log.SetLevel(log.InfoLevel)
}

// 初始化，为使用读写文件系统打开镜像文件，设置日志
func InitServe() {
	var err error
	fsfd, err = os.OpenFile(FS_IMG_FILE, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Read fs image error: ", err)
	}
	log.SetLevel(log.InfoLevel)
}

// 从二进制切片中读取对象
func readObject(buf []byte, ptrObject interface{}) {
	err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, ptrObject)
	if err != nil {
		panic(err)
	}
}

// 往二进制切片中写入等长的对象
func writeObject(buf []byte, object interface{}) {
	byteBuf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(byteBuf, binary.LittleEndian, object)
	if err != nil {
		panic(err)
	}
	copy(buf[:], byteBuf.Bytes())
}

// Block containing inode i
// 给出 index, 描述出index block对应的位置，SUPERBLOCK == 1
func IBLOCK(i uint32) uint32 {
	return i/uint32(IPB) + 2
}

/**
Block containing bit for block b
ninodes means ninode index
b means bios (in )
*/
// 这个应该表示的是bitmap block 对应的位置, B表示的是第几个块, 对应的是哪个位置
func BBLOCK(b uint16, ninodes uint8) uint8 {

	// 本来应该是 + 2, 但是实际上这里至少有一个block会被INODES TABLE占用，所以 + 3
	return uint8(uint8(b/BPB) + uint8(uint32(ninodes)/uint32(IPB)) + 3)
}

// 将一定大小的 bdata 写入对应的扇区
func writeToBlockDIO(blockNum uint32, bdata []byte) {
	written, err := fsfd.WriteAt(bdata[:], int64(blockNum*BLOCK_SIZE))
	if err != nil {
		panic(err)
	}
	if written != BLOCK_SIZE {
		panic("Written size error")
	}
}

// 读取文件中的一整个块
func readBlockDIO(blockNum uint32) []byte {
	data := make([]byte, BLOCK_SIZE)
	readSize, err := fsfd.ReadAt(data, int64(blockNum*BLOCK_SIZE))
	if err != nil {
		panic(err)
	}
	if readSize != BLOCK_SIZE {
		log.Fatalf("read size is %d and not blocksize", readSize)
	}
	return data
}

// 向文件中写入 inode
func fsyncINode(node *INode) {
	inodeBlockPos := IBLOCK(uint32(node.num))

	imap := readBlockDIO(inodeBlockPos)

	privateIndex := int(node.num) % int(IPB)

	begPos := privateIndex * int(unsafe.Sizeof(Dinode{}))
	endPos := begPos + int(unsafe.Sizeof(Dinode{}))
	log.Infof("fSync INode %d->%d in block %d", begPos, endPos, inodeBlockPos)
	//var dinode Dinode
	buf := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(buf, binary.LittleEndian, node.dinodeData)
	if err != nil {
		panic(err)
	}
	copy(imap[begPos:endPos], buf.Bytes())
	writeToBlockDIO(inodeBlockPos, imap)
	//thisINode := dinode.toINode()
}

// 向文件系统写入bytes
func writeFS(buf []byte, sec uint32) {
	// 直接写入 block
	off, err := fsfd.Seek(int64(BLOCK_SIZE)*int64(sec), 0)
	if err != nil {
		panic(err)
	}
	if off != int64(BLOCK_SIZE*sec) {
		panic("size error")
	}

	if len(buf) != BLOCK_SIZE {
		writeBuf := make([]byte, BLOCK_SIZE)
		if len(buf) > BLOCK_SIZE {
			log.Fatalf("Write buf size is %d, too large\n", len(buf))
		}
		copy(writeBuf, buf)
		buf = writeBuf
	}

	outsize, err := fsfd.WriteAt(buf, off)
	// || outsize != BLOCK_SIZE ban
	if err != nil {
		log.Fatalf("Only print %d\n", outsize)
	}
}
