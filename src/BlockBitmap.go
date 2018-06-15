package src

type BlockBitmap [BLOCK_SIZE]byte


func (bitmap *BlockBitmap) valid(index uint16) bool {
	bitmapIndex := index >> 3
	return (bitmap[int(bitmapIndex)] >> (index % 8)) & 1 == 1
}

func (bitmap *BlockBitmap) setValid(index uint16)  {
	bitmap.setBitmap(index, true)
}

func (bitmap *BlockBitmap) setInvalid(index uint16) {
	bitmap.setBitmap(index, false)
}

func (bitmap *BlockBitmap) setBitmap(index uint16, value bool) {
	bitmapIndex := index >> 3
	if value {
		bitmap[int(bitmapIndex)] |= 1 << (index % 8)
	} else {
		bitmap[int(bitmapIndex)] &= 255 & (^(1 << (index % 8)))
	}

}