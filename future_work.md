# Future Work

This document outlines potential enhancements to the memory manager beyond the current proof-of-concept.

## Thread Safety

The current implementation is entirely single-threaded. `Manager`, `store.Store`, `FreeList`, and `vmm.Table` share no synchronization primitives, so concurrent calls to `Alloc`, `Free`, `Read`, or `Write` from multiple goroutines will produce data races.

A straightforward first step is to add a `sync.RWMutex` at the `Manager` level: write-lock for `Alloc` and `Free` (which mutate both the free list and the VMM table), and read-lock for `Read` and `Write` (which only traverse the VMM table before touching the backing store). Finer-grained locking — a separate mutex per layer (`FreeList`, `vmm.Table`, `Disk`) — would reduce contention if read-heavy workloads dominate, at the cost of more complex lock ordering and potential deadlock surface.

The handle generator in `vmm.Table` (`nextHandle`) also needs atomic increment or mutex protection before it is safe to call `AddMapping` from concurrent goroutines.

## More Efficient `Reserve` (Avoid Slice Copying)

`FreeList.Reserve` currently removes a fully consumed extent from the middle of the `free` slice using `append(f.free[:idx], f.free[idx+1:]...)`. This is an O(n) copy of every element after `idx` on every full-extent consumption.

A more efficient approach is to replace the `[]Extent` slice with an intrusive doubly-linked list or a balanced BST (e.g., an interval tree) ordered by offset. Removal from a linked list is O(1) once the node pointer is in hand, and an interval tree supports O(log n) first-fit and best-fit searches simultaneously. For the common case where the free list stays short (few large extents), even replacing the linear scan with a simple sentinel-linked list would eliminate the repeated copy entirely.

An intermediate option that avoids a full data-structure overhaul is a "lazy deletion" tombstone scheme: mark a consumed extent as `Length == 0` in place and compact the slice only when the number of tombstones exceeds a threshold. This keeps the slice stable in memory and removes the per-removal copy at the cost of occasional O(n) compaction passes.

## Background Coalescing

`FreeList.Release` currently inserts returned extents and immediately runs `coalesce()`. That keeps the free list tidy, but it also pushes all merge work directly into the latency path of `Free`.

One alternative is to make `Free` cheap and defer coalescing to a background maintenance pass. In that design, `Release` would only validate and insert extents, while a separate goroutine periodically merges adjacent ranges or runs when fragmentation crosses a threshold. This shifts work away from the caller-facing path and can materially reduce tail latency for workloads with frequent frees.

The tradeoff is that allocation becomes more expensive while the free list remains fragmented. `Reserve` may need to inspect more extents before it can satisfy a request, and large allocations could fail transiently until the background coalescer catches up. A practical design therefore needs clear triggering policy: periodic compaction, threshold-based compaction, or a hybrid where `Reserve` can synchronously force a merge pass when a large request cannot be met.

If thread safety is added, background coalescing also affects locking strategy. A coalescer that mutates the free list concurrently with `Reserve` and `Release` requires either a coarse write lock around the entire structure or a more careful concurrent data-structure design.

## Memory Copy on `Read`

Every call to `Manager.Read` incurs at least two copies of the requested bytes:

1. `Disk.ReadAt` allocates a fresh `[]byte` and copies bytes out of the backing `[]byte` array into it.
2. `Manager.Read` then `append`s each per-span chunk into a second accumulator slice `out`.

For a fragmented allocation that spans k extents, the caller receives data that has been touched 2k times in total.

There are a few directions to reduce this overhead:

- **Single-allocation path**: when the logical range maps to exactly one contiguous physical span, return a sub-slice of the backing array directly (a zero-copy read). This requires careful ownership semantics — the slice must be treated as read-only and the caller must not retain it past the next mutating operation.
- **Caller-supplied buffer (`ReadInto`)**: expose an alternative API `ReadInto(h Handle, off int, dst []byte) (int, error)` that writes directly into a buffer provided by the caller, eliminating the intermediate allocation inside `Disk.ReadAt` entirely. This matches the `io.ReaderAt` pattern and composes well with pooled buffers.
- **Memory-mapped backing store**: replace the `[]byte`-backed `Disk` with a memory-mapped file (`mmap`). Reads then become page faults rather than explicit copies, and the OS can satisfy them from its page cache without ever entering user space for the copy itself.
