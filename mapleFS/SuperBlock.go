package mapleFS

import (
	log "github.com/sirupsen/logrus"
	"unsafe"
)

type superblock struct {
	Size uint32		// size of blocks
	Nblocks uint32	// number of datablocks
	Ninodes uint32	// number of inodes
}

// 初始化传入的 SUPER BLOCK 指针
func readsb(unInitSptr *superblock) {

	datas := make([]byte, BLOCK_SIZE)
	readSize, err := fsfd.ReadAt(datas, BLOCK_SIZE * 1)
	if readSize != BLOCK_SIZE || err != nil {
		log.Fatalf("Only read %d\n", readSize)
	}

	readObject(datas[:unsafe.Sizeof(superblock{})], unInitSptr)
}
