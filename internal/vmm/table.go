package vmm

import (
	"errors"
	"fmt"

	"memory_manager_poc/internal/store"
)

var (
	ErrHandleNotFound = errors.New("handle not found")
	ErrOutOfBounds    = errors.New("range out of bounds")
)

// Segment maps a logical range to a physical range.
type Segment struct {
	LogicalStart   int
	PhysicalOffset int
	Length         int
}

type Table struct {
	// TODO: Right now we use a uint64 for handle generation.
	// For large number of handle allocations, this will lead to wraparound and potential reuse of handles,
	// which can cause bugs if old handles are still referenced.
	nextHandle uint64               // monotonically increasing handle generator
	mappings   map[uint64][]Segment // handle -> segments. Each handle is basically a unique uint64.
	sizes      map[uint64]int       // handle -> total size of the mapping
}

func NewTable() *Table {
	return &Table{
		nextHandle: 1,
		mappings:   make(map[uint64][]Segment),
		sizes:      make(map[uint64]int),
	}
}

func (t *Table) AddMapping(extents []store.Extent, size int) (uint64, error) {
	if size <= 0 {
		return 0, fmt.Errorf("size must be > 0")
	}

	logical := 0
	segments := make([]Segment, 0, len(extents))

	for _, e := range extents {
		if logical >= size {
			break
		}

		take := min(size-logical, e.Length)
		if take <= 0 {
			// technically shouldn't happen if extents are well-formed, but just in case
			return 0, fmt.Errorf("invalid extent with zero length")
		}

		segments = append(segments, Segment{
			LogicalStart:   logical,
			PhysicalOffset: e.Offset,
			Length:         take,
		})
		logical += take
	}

	if logical != size {
		return 0, fmt.Errorf("mapping size mismatch")
	}

	h := t.nextHandle // select the next handle value
	t.nextHandle++    // increment for the next allocation

	t.mappings[h] = segments // store the segments for this handle
	t.sizes[h] = size        // store the total size for this handle
	return h, nil
}

func (t *Table) Remove(handle uint64) ([]store.Extent, error) {
	segments, ok := t.mappings[handle]
	if !ok {
		return nil, ErrHandleNotFound
	}

	delete(t.mappings, handle)
	delete(t.sizes, handle)

	extents := make([]store.Extent, 0, len(segments))
	for _, s := range segments {
		extents = append(extents, store.Extent{Offset: s.PhysicalOffset, Length: s.Length})
	}
	return extents, nil
}

func (t *Table) Lookup(handle uint64) ([]Segment, int, error) {
	segments, ok := t.mappings[handle]
	if !ok {
		return nil, 0, ErrHandleNotFound
	}
	size := t.sizes[handle]
	out := make([]Segment, len(segments))
	copy(out, segments)
	return out, size, nil
}

func (t *Table) ActiveHandles() int {
	return len(t.mappings)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
