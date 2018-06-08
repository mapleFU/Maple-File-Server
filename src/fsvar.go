package src

import "sync"

/**
inode is dinode in memory

    FS records file info in an "inode" on disk
    FS refers to inode with i-number (internal version of FD)
    inode must have link count (tells us when to free)
    inode must have count of open FDs
    inode deallocation deferred until last link and FD are gone

 */
type inode struct {
	num uint32	// 对应的序号
	ref int	// 引用计数
	lock sync.Mutex	// 内容的锁，暂时不会用到
	valid int32	// 是否在disk中被读出

	// copy of disk inode, 指向真实的block信息
	dinodeData dinode
}


// 析构函数
func (inode *inode) destruct()  {

}

// the system call to create inode
/**
遍历磁盘上的结构，寻找到空闲的结构，标注并返回

 */
func ialloc() *inode {
	return nil
}

// 遍历缓存找到对应的项
func iget() *inode {
	return nil
}


func test()  {

}