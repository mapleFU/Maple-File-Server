package src

import (
	"fmt"
	"testing"
)

var bitmap BlockBitmap

func TestBlockmap(t *testing.T)  {
	ptr := &bitmap
	fmt.Println(ptr.valid(213))
	ptr.setValid(213)
	fmt.Println(ptr.valid(213))
}
