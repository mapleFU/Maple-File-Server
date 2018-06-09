package src

import (
	"os"
	"log"
	"encoding/binary"
	"fmt"
	"bytes"

	"unsafe"
)

/**
生成fs 的包
 */

const FS_IMG_FILE = "maple-xv6.dmg"

var fsfd *os.File

func initfs()  {
	
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
	SIZE = 1024
	NINODES = 200
	NBLOCKS = 995
)

func writeFS(fs *os.File, buf []byte, sec uint32)  {
	// 直接写入 block
	off, err := fs.Seek(int64(BLOCK_SIZE) * int64(sec) , 0)
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

	outsize, err := fs.WriteAt(buf, off)
	// || outsize != BLOCK_SIZE ban
	if err != nil  {
		log.Fatalf("Only print %d\n", outsize)
	}
}

func GenerateFs()  {
	var err error
	// 基本的信息
	fsfd, err = os.OpenFile(FS_IMG_FILE, os.O_RDWR | os.O_CREATE | os.O_TRUNC,
		0666)
	if err != nil {
		log.Fatalln("cannot open the file")
	}
	defer fsfd.Close()

	sb := superblock{xuint32(1024), xuint32(200), xuint32(995)}
	fmt.Println(xuint32(1024), xuint32(200), xuint32(995))
	fmt.Println(sb)

	var bitblocks uint32 = SIZE / (BLOCK_SIZE * 8) + 1
	usedblocks := NINODES / uint32(IPB) + 3 + bitblocks
	freeblock := usedblocks

	fmt.Printf("used %d (bit %d ninode %d) free %d total %d\n", usedblocks,
		bitblocks, NINODES/uint32(IPB) + 1, freeblock, NBLOCKS+usedblocks)

	// make empty bytes
	emptyBlock := make([]byte, BLOCK_SIZE)
	//emptyBuffer := bytes.NewBuffer(emptyBlock)

	for i := uint32(0); i < NBLOCKS + usedblocks; i++ {
		// init physics block
		writeFS(fsfd, emptyBlock, i)
	}

	// init superblock
	buf := bytes.Buffer{}
	err = binary.Write(&buf, binary.LittleEndian, sb)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.Bytes())
	fmt.Println(len(buf.Bytes()) == int(unsafe.Sizeof(sb)))

	if err != nil {
		panic("bytes write error")
	}
	writeFS(fsfd, buf.Bytes(), 1)

	var sb2 superblock
	readsb(&sb2)
	fmt.Println(sb2)
	/** Unmarshal
	var sb2 superblock
	err = binary.Read(buf, binary.LittleEndian, &sb2)
	if err != nil {
		panic(err)
	}
	fmt.Println(sb2)
	 */


	// init inode map


	// init data bitmap

	// init root dir

}
