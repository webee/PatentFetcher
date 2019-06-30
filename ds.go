package main

import "sync"

type (
	// BitSet is bit set
	BitSet struct {
		sync.RWMutex
		bytes []byte
	}
)

// NewBitSet create a bit set.
func NewBitSet() *BitSet {
	return &BitSet{}
}

// NewBitSetFromBytes create a bit set from bytes.
func NewBitSetFromBytes(bs []byte) *BitSet {
	bytes := make([]byte, len(bs))
	copy(bytes, bs)
	return &BitSet{
		bytes: bytes,
	}
}

// Bytes get the underline bytes
func (s *BitSet) Bytes() []byte {
	bs := make([]byte, len(s.bytes))
	copy(bs, s.bytes)
	return bs
}

// Has x
func (s *BitSet) Has(x int) bool {
	s.RLock()
	defer s.RUnlock()
	n, b := x/8, byte(x%8)
	return n < len(s.bytes) && s.bytes[n]&(1<<b) != 0
}

// Add x
func (s *BitSet) Add(x int) {
	s.Lock()
	defer s.Unlock()
	n, b := x/8, byte(x%8)
	for n >= len(s.bytes) {
		s.bytes = append(s.bytes, 0)
	}
	s.bytes[n] |= 1 << b
}

// Del x
func (s *BitSet) Del(x int) {
	s.Lock()
	defer s.Unlock()
	n, b := x/8, byte(x%8)
	if n >= len(s.bytes) {
		return
	}
	s.bytes[n] &= 0xff ^ (1 << b)
}

// Union t
func (s *BitSet) Union(t *BitSet) {
	s.Lock()
	defer s.Unlock()
	for i, tb := range t.bytes {
		if i < len(s.bytes) {
			s.bytes[i] |= tb
		} else {
			s.bytes = append(s.bytes, tb)
		}
	}
}

// Len value count
func (s *BitSet) Len() int {
	s.RLock()
	defer s.RUnlock()
	n := 0
	for _, b := range s.bytes {
		for i := 0; i < 8; i++ {
			if b&0x01 != 0 {
				n++
			}
			b >>= 1
		}
	}
	return n
}

// MinNotExistsFrom value
func (s *BitSet) MinNotExistsFrom(f int) int {
	n, m := f/8, byte(f%8)
	s.RLock()
	defer s.RUnlock()
	for i := n; i < len(s.bytes); i++ {
		b := s.bytes[i]
		b >>= m
		if b != 0xff>>m {
			for b&0x01 != 0 {
				b >>= 1
				m++
			}
			return n*8 + int(m)
		}
		n++
		m = 0
	}
	return n*8 + int(m)
}
