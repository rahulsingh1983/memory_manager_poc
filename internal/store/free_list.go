package store

import (
	"errors"
	"fmt"
)

var (
	ErrInsufficientSpace = errors.New("insufficient physical space")
	ErrInvalidExtent     = errors.New("invalid extent")
)

// Extent is a contiguous physical range.
type Extent struct {
	Offset int
	Length int
}

type FreeList struct {
	total int
	free  []Extent
}

func NewFreeList(total int) (*FreeList, error) {
	if total <= 0 {
		return nil, fmt.Errorf("total must be > 0")
	}

	// Initialize with a single free extent covering the entire store.
	return &FreeList{
		total: total,
		free:  []Extent{{Offset: 0, Length: total}},
	}, nil
}

// Reserve takes bytes from free extents in first-fit order.
func (f *FreeList) Reserve(size int) ([]Extent, error) {
	if size <= 0 {
		return nil, ErrInvalidExtent
	}

	remaining := size
	reserved := make([]Extent, 0)

	for remaining > 0 {
		idx := -1
		for i := range f.free {
			if f.free[i].Length > 0 {
				// found a free extent, take from it
				idx = i
				break
			}
		}

		if idx == -1 {
			break
		}

		current := f.free[idx]

		if remaining < current.Length {
			// Partial take: reserve from the head of this extent and keep the tail free.
			reserved = append(reserved, Extent{Offset: current.Offset, Length: remaining})
			f.free[idx].Offset += remaining
			f.free[idx].Length -= remaining
			remaining = 0
			continue
		}

		// Full take: consume this entire extent and remove it from the free list.
		reserved = append(reserved, current)
		remaining -= current.Length
		f.free = append(f.free[:idx], f.free[idx+1:]...)
	}

	if remaining > 0 {
		// Roll back partial reservation.
		_ = f.Release(reserved)
		return nil, ErrInsufficientSpace
	}

	return reserved, nil
}

// Release returns extents to the free list and coalesces adjacent ranges.
func (f *FreeList) Release(extents []Extent) error {
	for _, e := range extents {
		if err := f.validateExtent(e); err != nil {
			return err
		}
		if err := f.insertExtent(e); err != nil {
			return err
		}
	}

	f.coalesce()
	return nil
}

func (f *FreeList) Snapshot() []Extent {
	out := make([]Extent, len(f.free))
	copy(out, f.free)
	return out
}

func (f *FreeList) FreeBytes() int {
	total := 0
	for _, e := range f.free {
		total += e.Length
	}
	return total
}

func (f *FreeList) validateExtent(e Extent) error {
	if e.Length <= 0 || e.Offset < 0 || e.Offset+e.Length > f.total {
		return ErrInvalidExtent
	}
	return nil
}

func (f *FreeList) insertExtent(newExtent Extent) error {
	insertAt := len(f.free)
	// This function assumes that the extents are in sorted order by offset.
	// We find the correct insertion point for the new extent to maintain this order.
	for i, curr := range f.free {
		if overlaps(curr, newExtent) {
			return ErrInvalidExtent
		}
		if newExtent.Offset < curr.Offset {
			insertAt = i
			break
		}
	}

	f.free = append(f.free, Extent{})
	// make space for the new extent by shifting elements to the right
	copy(f.free[insertAt+1:], f.free[insertAt:])
	f.free[insertAt] = newExtent
	return nil
}

func (f *FreeList) coalesce() {
	if len(f.free) <= 1 {
		// nothing to coalesce
		return
	}

	// Create a brand new free list by merging adjacent extents.
	// This is simpler than trying to coalesce in place.
	// TODO: if this becomes a bottleneck,
	// we can optimize by only coalescing around the newly inserted extents.
	merged := make([]Extent, 0, len(f.free))
	current := f.free[0]

	for i := 1; i < len(f.free); i++ {
		next := f.free[i]
		if current.Offset+current.Length == next.Offset {
			current.Length += next.Length
			continue
		}

		merged = append(merged, current)
		current = next
	}

	merged = append(merged, current)
	f.free = merged
}

func overlaps(a, b Extent) bool {
	aEnd := a.Offset + a.Length
	bEnd := b.Offset + b.Length
	return a.Offset < bEnd && b.Offset < aEnd
}

type Store struct {
	disk     *Disk
	freeList *FreeList
}

func New(size int) (*Store, error) {
	disk, err := NewDisk(size)
	if err != nil {
		return nil, err
	}

	freeList, err := NewFreeList(size)
	if err != nil {
		return nil, err
	}

	return &Store{disk: disk, freeList: freeList}, nil
}

func (s *Store) Reserve(size int) ([]Extent, error) {
	return s.freeList.Reserve(size)
}

func (s *Store) Release(extents []Extent) error {
	return s.freeList.Release(extents)
}

func (s *Store) ReadAt(offset, n int) ([]byte, error) {
	return s.disk.ReadAt(offset, n)
}

func (s *Store) WriteAt(offset int, in []byte) error {
	return s.disk.WriteAt(offset, in)
}

func (s *Store) FreeBytes() int {
	return s.freeList.FreeBytes()
}

func (s *Store) TotalBytes() int {
	return s.disk.Len()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
