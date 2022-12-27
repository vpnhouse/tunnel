package statutils

import (
	"math/rand"
	"testing"
)

func TestAvgValue(t *testing.T) {
	avgVal := NewAvgValue(10)
	min := 10
	max := 30
	for i := 0; i < 100; i++ {
		v := int64(rand.Intn(max-min) + min)
		avg := avgVal.Push(v)
		t.Logf("%v -> %v <- (%v)", v, avg, avgVal.values)
	}
}

func TestAvgValueOrdinal(t *testing.T) {
	avgVal := NewAvgValue(10)
	for i := 0; i < 100; i++ {
		v := int64(i)
		avg := avgVal.Push(v)
		t.Logf("%v -> %v <- (%v)", v, avg, avgVal.values)
	}
}
