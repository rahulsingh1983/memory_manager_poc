package vmm

import "fmt"

type PhysicalSpan struct {
	PhysicalOffset int
	Length         int
}

func (t *Table) Translate(handle uint64, off, n int) ([]PhysicalSpan, error) {
	if off < 0 || n < 0 {
		return nil, ErrOutOfBounds
	}
	if n == 0 {
		return []PhysicalSpan{}, nil
	}

	segments, size, err := t.Lookup(handle)
	if err != nil {
		return nil, err
	}

	if off+n > size {
		return nil, ErrOutOfBounds
	}

	remaining := n
	cursor := off
	out := make([]PhysicalSpan, 0, len(segments))

	for _, seg := range segments {
		segStart := seg.LogicalStart
		segEnd := seg.LogicalStart + seg.Length

		if cursor >= segEnd {
			continue
		}

		if cursor < segStart {
			// This should never happen if the segments are well-formed and non-overlapping.
			// Segments should always be of the form [0, x), [x, y), [y, z) with no gaps.
			// If we encounter a gap, it means the mapping is malformed.
			return nil, fmt.Errorf("mapping gap at logical offset %d", cursor)
		}

		// if we reach here, mean that cursor is within the current segment.
		// We need to calculate how much of the segment we can take.

		start := max(cursor, segStart)
		take := min(remaining, segEnd-start)

		out = append(out, PhysicalSpan{
			PhysicalOffset: seg.PhysicalOffset + (start - segStart),
			Length:         take,
		})

		cursor += take
		remaining -= take
		if remaining == 0 {
			break
		}
	}

	if remaining != 0 {
		return nil, ErrOutOfBounds
	}

	return out, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
