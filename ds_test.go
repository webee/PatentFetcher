package main

import (
	"testing"
)

func TestBitSet_Has(t *testing.T) {
	bs := NewBitSet()
	bs.Add(1)
	bs.Add(3)
	bs.Add(5)
	bs.Add(7)
	bs.Add(100)
	bs.Add(101)
	bs.Add(102)
	bs.Add(103)
	bs.Add(104)
	bs.Add(105)
	bs.Add(106)
	bs.Del(101)
	bs.Del(103)
	bs.Del(105)
	tests := []struct {
		val  int
		want bool
	}{
		{0, false}, {1, true}, {2, false}, {3, true}, {4, false}, {5, true}, {6, false}, {7, true},
		{8, false}, {9, false},
		{100, true}, {101, false}, {102, true}, {103, false}, {104, true}, {105, false}, {106, true}, {107, false},
	}
	for _, tt := range tests {
		if got := bs.Has(tt.val); got != tt.want {
			t.Errorf("BitSet.Has(%d) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestBitSet_Len(t *testing.T) {
	bs := NewBitSet()
	if got := bs.Len(); got != 0 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 0)
	}
	bs.Add(123)
	if got := bs.Len(); got != 1 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 1)
	}
	bs.Add(8)
	if got := bs.Len(); got != 2 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 2)
	}
	bs.Del(7)
	if got := bs.Len(); got != 2 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 2)
	}
	bs.Del(123)
	if got := bs.Len(); got != 1 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 1)
	}
	bs.Add(9999)
	if got := bs.Len(); got != 2 {
		t.Errorf("BitSet.Len() = %v, want %v", got, 2)
	}
}

func TestBitSet_MinNotExistsFrom(t *testing.T) {
	bs := NewBitSet()
	if got := bs.MinNotExistsFrom(0); got != 0 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 0, got, 0)
	}
	if got := bs.MinNotExistsFrom(1); got != 1 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 1, got, 1)
	}
	if got := bs.MinNotExistsFrom(123); got != 123 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 123, got, 123)
	}

	for i := 30; i <= 70; i++ {
		bs.Add(i)
	}
	if got := bs.MinNotExistsFrom(0); got != 0 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 0, got, 0)
	}
	if got := bs.MinNotExistsFrom(1); got != 1 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 1, got, 1)
	}
	if got := bs.MinNotExistsFrom(33); got != 71 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 33, got, 71)
	}
	if got := bs.MinNotExistsFrom(123); got != 123 {
		t.Errorf("BitSet.MinNotExistsFrom(%d) = %v, want %v", 123, got, 123)
	}
}
