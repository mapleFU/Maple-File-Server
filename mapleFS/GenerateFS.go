package mapleFS

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"unsafe"
)

func GenerateFs() {
	initMkfs()
	defer fsfd.Close()

	sb := superblock{xuint32(1024), xuint32(200), xuint32(995)}
	fmt.Println(xuint32(1024), xuint32(200), xuint32(995))
	fmt.Println(sb)

	var bitblocks uint32 = SIZE/(BLOCK_SIZE*8) + 1
	usedblocks := NINODES/uint32(IPB) + 3 + bitblocks

	freeblock := usedblocks

	fmt.Printf("used %d (bit %d ninode %d) free %d total %d\n", usedblocks,
		bitblocks, NINODES/uint32(IPB)+1, freeblock, NBLOCKS+usedblocks)

	// make empty bytes
	emptyBlock := make([]byte, BLOCK_SIZE)
	//emptyBuffer := bytes.NewBuffer(emptyBlock)

	for i := uint32(0); i < NBLOCKS+usedblocks; i++ {
		// init physics block
		writeToBlockDIO(i, emptyBlock)
		//writeFS(emptyBlock, i)
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

	copy(emptyBlock, buf.Bytes())
	writeToBlockDIO(1, emptyBlock)

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
	writeDi := Dinode{Size: MAX_UINT32}
	DINODE_SIZE := int(unsafe.Sizeof(writeDi))
	for i := lowerB; i <= upperB; i++ {
		//log.WithField("event", "init").Infof("init dinodes %d", i)
		// 读取对应的block
		bytesData := make([]byte, BLOCK_SIZE)

		for j := 0; j < int(IPB); j++ {
			buf := bytes.Buffer{}
			err := binary.Write(&buf, binary.LittleEndian, writeDi)
			if err != nil {
				panic(err)
			}
			bufData := buf.Bytes()
			if len(bufData) != DINODE_SIZE {
				log.Fatalf("bufData size is not DINODE_SIZE")
			}
			copy(bytesData[j*DINODE_SIZE:(j+1)*DINODE_SIZE], bufData)
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
