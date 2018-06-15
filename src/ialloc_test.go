package src

import (
	"testing"
	"github.com/sirupsen/logrus"
)

func TestIalloc(t *testing.T) {
	initServe()
	logrus.SetLevel(logrus.DebugLevel)
	i1 := ialloc()
	i2 := ialloc()
	logrus.Debug(i1.num, " -- ", i2.num)
	if i1.num == i2.num {
		t.Fatalf("Test Error")
	}
}