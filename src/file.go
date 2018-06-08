package src



/**
struct file {
	enum { FD_NONE, FD_PIPE, FD_INODE } type;
	int ref; // reference count
	char readable;
	char writable;
	struct pipe *pipe;
	struct inode *ip;
	uint off;
};
 */

 // 文件类型，这里有用么？
type FileType uint

// https://stackoverflow.com/questions/14426366/what-is-an-idiomatic-way-of-representing-enums-in-go
const (
	FD_NONE = iota
	FD_PIPE = iota
	FD_INODE = iota
)

type fsfile struct {
	ref int	// ref cnt
	inodePtr *inode
	fileType FileType

	readable bool
	writeable bool
}

func createFile() *fsfile {
	panic("Not")
}