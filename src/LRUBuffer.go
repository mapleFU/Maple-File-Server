package src

import "container/list"

type LRUBuffer struct {
	bufferList *list.List	// 存储链表
	maxSize int	// 最大的长度
}

func NewLRUBuf(lruSize int) *LRUBuffer {
	var lruBuf LRUBuffer
	lruBuf.bufferList = list.New()
	lruBuf.maxSize = lruSize
	return &lruBuf
}

func (lruBuf *LRUBuffer) Len() int {
	return lruBuf.bufferList.Len()
}

func (lruBuf *LRUBuffer) Evict()  {

}

func (lruBuf *LRUBuffer) Add() *list.Element {
	unimpletedError()
	return nil
}

func (lruBuf *LRUBuffer) Find()  {

}