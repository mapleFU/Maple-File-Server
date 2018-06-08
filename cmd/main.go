package main

import (
	"github.com/mapleFU/TongjiFileLab/src"

	"fmt"
	"unsafe"
)

func main()  {
	fmt.Println(uint8(src.IPB), src.IPB)
	//var s1 [4]byte
	//fmt.Println(unsafe.Sizeof(s1))
	//var s2 [16]byte
	//fmt.Println(unsafe.Sizeof(s2))
	type S1 struct {
		s1 []byte	// slice is fixed size
	}

	type S2 struct {
		s [4]byte
	}

	type S3 struct {
		s [16]byte
	}

	fmt.Println(unsafe.Sizeof(S1{}), unsafe.Sizeof(S2{}), unsafe.Sizeof(S3{}))

	fmt.Println(src.IPB, src.MAX_FILE, src.DIRENT_SIZE)
}
