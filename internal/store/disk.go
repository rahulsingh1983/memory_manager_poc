package store

import "fmt"

// Disk is a byte-array-backed physical store.
type Disk struct {
	data []byte
}

func NewDisk(size int) (*Disk, error) {
	if size <= 0 {
		return nil, fmt.Errorf("disk size must be > 0")
	}

	return &Disk{data: make([]byte, size)}, nil
}

func (d *Disk) Len() int {
	return len(d.data)
}

func (d *Disk) ReadAt(offset, n int) ([]byte, error) {
	if offset < 0 || n < 0 || offset+n > len(d.data) {
		return nil, fmt.Errorf("read out of bounds")
	}

	out := make([]byte, n)
	copy(out, d.data[offset:offset+n])
	return out, nil
}

func (d *Disk) WriteAt(offset int, in []byte) error {
	if offset < 0 || offset+len(in) > len(d.data) {
		return fmt.Errorf("write out of bounds")
	}

	copy(d.data[offset:], in)
	return nil
}