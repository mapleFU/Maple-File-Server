package mapleFS

type FileType uint

// https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
const (
	FD_NONE  = iota
	FD_PIPE  = iota
	FD_INODE = iota
)

type FsFile struct {
	ref      int // ref cnt
	inodePtr *inode
	fileType FileType

	readable  bool
	writeable bool
}

func createFile() *FsFile {
	unimpletedError()
	return nil
}
