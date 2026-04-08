package memory

import (
	"fmt"

	"memory_manager_poc/internal/store"
	"memory_manager_poc/internal/vmm"
)

type Manager struct {
	store *store.Store
	vmm   *vmm.Table
}

func New(cfg Config) (*Manager, error) {
	if cfg.DiskSize <= 0 {
		return nil, ErrInvalidSize
	}

	if cfg.PlacementStrategy == "" {
		cfg.PlacementStrategy = PlacementFirstFit
	}

	if cfg.PlacementStrategy != PlacementFirstFit {
		return nil, fmt.Errorf("unsupported placement strategy: %q", cfg.PlacementStrategy)
	}

	st, err := store.New(cfg.DiskSize)
	if err != nil {
		return nil, err
	}

	return &Manager{store: st, vmm: vmm.NewTable()}, nil
}

func (m *Manager) Alloc(size int) (Handle, error) {
	if size <= 0 {
		return Handle{}, ErrInvalidSize
	}

	extents, err := m.store.Reserve(size)
	if err != nil {
		if err == store.ErrInsufficientSpace {
			return Handle{}, ErrOutOfMemory
		}
		return Handle{}, err
	}

	h, err := m.vmm.AddMapping(extents, size)
	if err != nil {
		_ = m.store.Release(extents)
		return Handle{}, err
	}

	return Handle{id: h}, nil
}

func (m *Manager) Free(h Handle) error {
	if h.id == 0 {
		return ErrInvalidHandle
	}

	extents, err := m.vmm.Remove(h.id)
	if err != nil {
		if err == vmm.ErrHandleNotFound {
			// In v1, not found also covers double free semantics.
			return ErrInvalidHandle
		}
		return err
	}

	if err := m.store.Release(extents); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Read(h Handle, off, n int) ([]byte, error) {
	if h.id == 0 {
		return nil, ErrInvalidHandle
	}
	if off < 0 || n < 0 {
		return nil, ErrOutOfBounds
	}
	if n == 0 {
		return []byte{}, nil
	}

	spans, err := m.vmm.Translate(h.id, off, n)
	if err != nil {
		if err == vmm.ErrHandleNotFound {
			return nil, ErrInvalidHandle
		}
		if err == vmm.ErrOutOfBounds {
			return nil, ErrOutOfBounds
		}
		return nil, err
	}

	out := make([]byte, 0, n)
	for _, span := range spans {
		chunk, err := m.store.ReadAt(span.PhysicalOffset, span.Length)
		if err != nil {
			return nil, err
		}
		out = append(out, chunk...)
	}

	return out, nil
}

func (m *Manager) Write(h Handle, off int, in []byte) error {
	if h.id == 0 {
		return ErrInvalidHandle
	}
	if off < 0 {
		return ErrOutOfBounds
	}
	if len(in) == 0 {
		return nil
	}

	spans, err := m.vmm.Translate(h.id, off, len(in))
	if err != nil {
		if err == vmm.ErrHandleNotFound {
			return ErrInvalidHandle
		}
		if err == vmm.ErrOutOfBounds {
			return ErrOutOfBounds
		}
		return err
	}

	written := 0
	for _, span := range spans {
		chunk := in[written : written+span.Length]
		if err := m.store.WriteAt(span.PhysicalOffset, chunk); err != nil {
			return err
		}
		written += span.Length
	}

	return nil
}

func (m *Manager) Stats() Stats {
	total := m.store.TotalBytes()
	free := m.store.FreeBytes()
	used := total - free
	return Stats{
		TotalBytes:    total,
		UsedBytes:     used,
		FreeBytes:     free,
		ActiveHandles: m.vmm.ActiveHandles(),
	}
}