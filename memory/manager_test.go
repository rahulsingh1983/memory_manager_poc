package memory

import "testing"

func TestAllocAndReadWriteAcrossNonContiguousExtents(t *testing.T) {
	mgr, err := New(Config{DiskSize: 10})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	h1, err := mgr.Alloc(4)
	if err != nil {
		t.Fatalf("Alloc(4) failed: %v", err)
	}
	h2, err := mgr.Alloc(3)
	if err != nil {
		t.Fatalf("Alloc(3) failed: %v", err)
	}
	h3, err := mgr.Alloc(3)
	if err != nil {
		t.Fatalf("Alloc(3) failed: %v", err)
	}

	if err := mgr.Free(h1); err != nil {
		t.Fatalf("Free(h1) failed: %v", err)
	}
	if err := mgr.Free(h3); err != nil {
		t.Fatalf("Free(h3) failed: %v", err)
	}

	// There is no contiguous extent of length 5, but total free bytes are enough.
	h4, err := mgr.Alloc(5)
	if err != nil {
		t.Fatalf("Alloc(5) should succeed via multiple extents: %v", err)
	}

	payload := "hello"
	if err := mgr.Write(h4, 0, []byte(payload)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	out, err := mgr.Read(h4, 0, len(payload))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(out) != payload {
		t.Fatalf("unexpected payload: got %q want %q", string(out), payload)
	}

	// Keep h2 alive to ensure allocation state stays valid.
	_ = h2
}

func TestInvalidAndDoubleFree(t *testing.T) {
	mgr, err := New(Config{DiskSize: 8})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	h, err := mgr.Alloc(4)
	if err != nil {
		t.Fatalf("Alloc failed: %v", err)
	}

	if err := mgr.Free(h); err != nil {
		t.Fatalf("first Free failed: %v", err)
	}

	if err := mgr.Free(h); err != ErrInvalidHandle {
		t.Fatalf("expected ErrInvalidHandle on double free, got %v", err)
	}
}

func TestReadWriteBounds(t *testing.T) {
	mgr, err := New(Config{DiskSize: 6})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	h, err := mgr.Alloc(4)
	if err != nil {
		t.Fatalf("Alloc failed: %v", err)
	}

	if err := mgr.Write(h, 3, []byte("ab")); err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, got %v", err)
	}

	if _, err := mgr.Read(h, 3, 2); err != ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, got %v", err)
	}
}

func TestAllocFreeReallocScattered(t *testing.T) {
	mgr, err := New(Config{DiskSize: 5})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	b1, err := mgr.Alloc(1)
	if err != nil {
		t.Fatalf("Alloc b1 failed: %v", err)
	}
	b2, err := mgr.Alloc(1)
	if err != nil {
		t.Fatalf("Alloc b2 failed: %v", err)
	}
	b3, err := mgr.Alloc(1)
	if err != nil {
		t.Fatalf("Alloc b3 failed: %v", err)
	}
	b4, err := mgr.Alloc(1)
	if err != nil {
		t.Fatalf("Alloc b4 failed: %v", err)
	}
	b5, err := mgr.Alloc(1)
	if err != nil {
		t.Fatalf("Alloc b5 failed: %v", err)
	}

	if err := mgr.Write(b1, 0, []byte("a")); err != nil {
		t.Fatalf("Write b1 failed: %v", err)
	}
	out, err := mgr.Read(b1, 0, 1)
	if err != nil {
		t.Fatalf("Read b1 failed: %v", err)
	}
	if string(out) != "a" {
		t.Fatalf("b1: got %q want %q", string(out), "a")
	}

	if err := mgr.Write(b2, 0, []byte("z")); err != nil {
		t.Fatalf("Write b2 failed: %v", err)
	}
	out, err = mgr.Read(b2, 0, 1)
	if err != nil {
		t.Fatalf("Read b2 failed: %v", err)
	}
	if string(out) != "z" {
		t.Fatalf("b2: got %q want %q", string(out), "z")
	}

	if err := mgr.Free(b2); err != nil {
		t.Fatalf("Free b2 failed: %v", err)
	}
	if err := mgr.Free(b4); err != nil {
		t.Fatalf("Free b4 failed: %v", err)
	}

	// Freed handles should no longer be accessible.
	if _, err := mgr.Read(b2, 0, 1); err != ErrInvalidHandle {
		t.Fatalf("expected ErrInvalidHandle reading freed b2, got %v", err)
	}
	if _, err := mgr.Read(b4, 0, 1); err != ErrInvalidHandle {
		t.Fatalf("expected ErrInvalidHandle reading freed b4, got %v", err)
	}

	// 2 bytes freed across non-contiguous slots; b6 should succeed.
	b6, err := mgr.Alloc(2)
	if err != nil {
		t.Fatalf("Alloc b6 failed: %v", err)
	}

	if err := mgr.Write(b6, 0, []byte("xy")); err != nil {
		t.Fatalf("Write b6 failed: %v", err)
	}
	out, err = mgr.Read(b6, 0, 2)
	if err != nil {
		t.Fatalf("Read b6 failed: %v", err)
	}
	if string(out) != "xy" {
		t.Fatalf("b6: got %q want %q", string(out), "xy")
	}

	// Keep surviving handles alive.
	_, _, _ = b1, b3, b5
}
