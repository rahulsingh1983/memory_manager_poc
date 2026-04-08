package store

import "testing"

func TestReserveSplitAndReleaseCoalesce(t *testing.T) {
	s, err := New(10)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	e1, err := s.Reserve(4)
	if err != nil {
		t.Fatalf("Reserve(4) failed: %v", err)
	}
	e2, err := s.Reserve(3)
	if err != nil {
		t.Fatalf("Reserve(3) failed: %v", err)
	}

	if err := s.Release(e1); err != nil {
		t.Fatalf("Release(e1) failed: %v", err)
	}
	if err := s.Release(e2); err != nil {
		t.Fatalf("Release(e2) failed: %v", err)
	}

	if free := s.FreeBytes(); free != 10 {
		t.Fatalf("unexpected free bytes: got %d want 10", free)
	}
}

func TestReserveAcrossFragmentedExtents(t *testing.T) {
	s, err := New(10)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	e1, _ := s.Reserve(4)
	_, _ = s.Reserve(3)
	e3, _ := s.Reserve(3)

	if err := s.Release(e1); err != nil {
		t.Fatalf("Release(e1) failed: %v", err)
	}
	if err := s.Release(e3); err != nil {
		t.Fatalf("Release(e3) failed: %v", err)
	}

	reserved, err := s.Reserve(5)
	if err != nil {
		t.Fatalf("Reserve(5) failed: %v", err)
	}
	if len(reserved) < 2 {
		t.Fatalf("expected fragmented reservation, got %+v", reserved)
	}

	total := 0
	for _, e := range reserved {
		total += e.Length
	}
	if total != 5 {
		t.Fatalf("unexpected total reserved: got %d want 5", total)
	}
}