package vmm

import (
	"testing"

	"memory_manager_poc/internal/store"
)

func TestTranslateAcrossSegments(t *testing.T) {
	table := NewTable()
	h, err := table.AddMapping([]store.Extent{{Offset: 0, Length: 2}, {Offset: 10, Length: 3}}, 5)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	spans, err := table.Translate(h, 1, 3)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[0].PhysicalOffset != 1 || spans[0].Length != 1 {
		t.Fatalf("unexpected first span: %+v", spans[0])
	}
	if spans[1].PhysicalOffset != 10 || spans[1].Length != 2 {
		t.Fatalf("unexpected second span: %+v", spans[1])
	}
}

func TestTranslateErrors(t *testing.T) {
	table := NewTable()
	h, err := table.AddMapping([]store.Extent{{Offset: 3, Length: 4}}, 4)
	if err != nil {
		t.Fatalf("AddMapping failed: %v", err)
	}

	if _, err := table.Translate(h, 3, 2); err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, got %v", err)
	}

	if _, err := table.Translate(999, 0, 1); err != ErrHandleNotFound {
		t.Fatalf("expected ErrHandleNotFound, got %v", err)
	}
}