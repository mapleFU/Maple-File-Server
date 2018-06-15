package src

import (
	"os"
	"fmt"
	"encoding/binary"
	"bytes"
	"unsafe"
	log "github.com/sirupsen/logrus"
)

/**
生成fs 的包
 */
const MAX_UINT32  = 4294967295
const FS_IMG_FILE = "maple-xv6.dmg"

var fsfd *os.File

func initMkfs()  {
	var err error
	// 基本的信息
	fsfd, err = os.OpenFile(FS_IMG_FILE, os.O_RDWR | os.O_CREATE | os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalln("cannot open the file")
	}
	log.SetLevel(log.InfoLevel)
}

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

const (
	// 总共大小
	SIZE = 1024
	// INODES 的数量
	NINODES = 200
	// BLOCKS 的数量
	NBLOCKS = 995
)


func writeFS(buf []byte, sec uint32)  {
	// 直接写入 block
	off, err := fsfd.Seek(int64(BLOCK_SIZE) * int64(sec) , 0)
	if err != nil {
		panic(err)
	}
	if off != int64(BLOCK_SIZE * sec) {
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
	if err != nil  {
		log.Fatalf("Only print %d\n", outsize)
	}
}

const BITMAP_BLOCK_NUM uint32 = SIZE / (BLOCK_SIZE * 8) + 1
var usedblocks uint32

func GenerateFs()  {
	initMkfs()
	defer fsfd.Close()

	sb := superblock{xuint32(1024), xuint32(200), xuint32(995)}
	fmt.Println(xuint32(1024), xuint32(200), xuint32(995))
	fmt.Println(sb)

	var bitblocks uint32 = SIZE / (BLOCK_SIZE * 8) + 1
	usedblocks = NINODES / uint32(IPB) + 3 + bitblocks

	freeblock := usedblocks

	fmt.Printf("used %d (bit %d ninode %d) free %d total %d\n", usedblocks,
		bitblocks, NINODES/uint32(IPB) + 1, freeblock, NBLOCKS+usedblocks)

	// make empty bytes
	emptyBlock := make([]byte, BLOCK_SIZE)
	//emptyBuffer := bytes.NewBuffer(emptyBlock)

	for i := uint32(0); i < NBLOCKS + usedblocks; i++ {
		// init physics block
		writeFS(emptyBlock, i)
	}

	// init superblock
	buf := bytes.Buffer{}
	err := binary.Write(&buf, binary.LittleEndian, sb)
	if err != nil {
		panic(err)
	}

	if len(buf.Bytes()) != int(unsafe.Sizeof(sb)) {
		log.Fatal("SuperBlock size error, got ", len(buf.Bytes()))
	}

	if err != nil {
		panic("bytes write error")
	}
	writeFS(buf.Bytes(), 1)

	// test sb
	var sb2 superblock
	readsb(&sb2)
	fmt.Println(sb2)
	// write bitmap --> 反正都它妈是0 --> 不对，前面都是1啊
	initBitBlock()

	// init inode --> 反正一个都没有
	lowerB := IBLOCK(0)
	upperB := IBLOCK(NINODES)
	log.Infof("INode init from %d to %d", lowerB, upperB)
	// 读取的dinode
	writeDi := Dinode{Size:MAX_UINT32}
	DINODE_SIZE := int(unsafe.Sizeof(writeDi))
	for i := lowerB; i <= upperB; i++ {
		//log.WithField("event", "init").Infof("init dinodes %d", i)
		// 读取对应的block
		bytesData := make([]byte, BLOCK_SIZE)

		for j := 0; j < int(IPB); j++  {
			buf := bytes.Buffer{}
			err := binary.Write(&buf, binary.LittleEndian, writeDi)
			if err != nil {
				panic(err)
			}
			bufData := buf.Bytes()
			if len(bufData) != DINODE_SIZE {
				log.Fatalf("bufData size is not DINODE_SIZE")
			}
			copy(bytesData[j*DINODE_SIZE: (j+1)*DINODE_SIZE], bufData)
		}
		writeToBlockDIO(i, bytesData)
	}

	// init root dir

	rootDir := MkRootDir()
	WalkDir(rootDir)
	log.Infof("Root INode: ", rootDir.num)
}


func initBitBlock() {
	// 这里的问题是我好像把bitmap写成了bytemap
	blockID := BBLOCK(0, NINODES)
	blockBytes := readBlockDIO(uint32(blockID))
	// TODO:为什么每次都要copy
	var bitmap BlockBitmap
	copy(bitmap[:], blockBytes)
	//bitmap := &BlockBitmap{barray}
	//log.Infof("Init block at %d", blockID)
	var index uint16
	for index = 0; index < uint16(usedblocks); index++ {
		bitmap.setValid(index)
	}
	log.Debug(bitmap)
	writeToBlockDIO(uint32(blockID), bitmap[:])
}