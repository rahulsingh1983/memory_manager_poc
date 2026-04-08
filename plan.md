## Plan: Initialize Go Memory Manager Scaffold

Create a library-first Go project scaffold in /Users/rahulsingh/Documents/memory_manager_poc by mirroring the useful structure from /Users/rahulsingh/Documents/sample_golang_app, while intentionally excluding all GRPC/HTTP/proto/generated layers. Focus on library APIs and unit-test-driven validation, and seed the domain around virtual memory allocation/free over a byte-array-backed simulated disk.

**Steps**
1. Baseline scaffold replication (*blocking*): create top-level project files aligned with the sample style but slimmed for this use case: go.mod, go.sum (if needed after mod tidy), .go-version, .gitignore, README.md, Makefile.
2. Create package layout (*depends on 1*): create a top-level public package (for allocator APIs), plus internal packages for storage and virtual-memory translation.
3. Define core domain contracts (*depends on 2*): add memory-manager types for allocation/free behavior and disk abstraction using a []byte simulation strategy.
4. Implement minimal storage layer skeleton (*depends on 3*): add internal store package with a byte-array disk model and extent-management placeholders.
5. Implement virtual mapping layer (*depends on 4*): add internal vmm package for handle-to-segment mappings and address translation.
6. Implement manager layer skeleton (*depends on 5*): add public manager APIs that delegate to store and vmm and return stable errors.
7. Add test skeletons (*depends on 4, 5, and 6*): add focused unit tests for allocation/free invariants, fragmented allocations, and translation edge cases.
8. Remove service-oriented concerns (*continuous check across steps*): do not include internal/grpcserver, internal/httpserver, internal/routes, proto, gen, or cmd equivalents.
9. Verify scaffold quality (*depends on 1-8*): run go test ./... and go vet ./... to ensure clean initialization.

**Relevant files**
- /Users/rahulsingh/Documents/sample_golang_app/go.mod — reference dependency and module conventions, then simplify for non-service app.
- /Users/rahulsingh/Documents/sample_golang_app/Makefile — reuse build/test conventions, excluding proto/grpc targets.
- /Users/rahulsingh/Documents/sample_golang_app/internal/store/memory.go — reference in-memory data modeling patterns.
- /Users/rahulsingh/Documents/sample_golang_app/internal/store/memory_test.go — reference unit test style for storage behavior.
- /Users/rahulsingh/Documents/sample_golang_app/internal/handlers/todo.go — reference business-layer separation pattern.
- /Users/rahulsingh/Documents/memory_manager_poc (workspace root) — target location for scaffold creation.

**Verification**
1. Confirm generated tree excludes service/proto folders by listing directories and checking for absence of grpcserver/httpserver/routes/proto/gen.
2. Run go mod tidy and ensure only necessary dependencies remain.
3. Run go test ./... and verify all initial tests pass.
4. Run go vet ./... and confirm no obvious static-analysis issues.
5. Run focused package tests for translation and fragmented allocation paths to validate non-contiguous virtual allocation behavior.

**Decisions**
- Chosen app shape: Library-first.
- Chosen construction style: No application entrypoint; unit tests are the demonstration mechanism.
- Included scope: scaffold + core package skeletons + starter tests.
- Excluded scope: HTTP endpoints, GRPC server/client/proto generation, persistence beyond in-memory byte array, advanced allocator algorithms.
- Concurrency model for v1: single-threaded implementation only; no mutexes, atomic operations, or thread-safety guarantees.
- Free ordering contract for v1: Manager must call VMM to remove/retrieve handle segment mappings before returning physical extents to Store for coalescing.



**Further Considerations**
1. Allocator strategy for v1: first-fit is recommended for initial correctness and simpler tests.
2. API style: use idiomatic Go names (Alloc/Free) while documenting malloc/free equivalence to avoid C-specific naming friction.
3. Demonstration approach: rely on unit tests only; no executable demo target in v1.

**Planned Go file contents (virtualized, non-contiguous allocations)**
1. memory/manager.go
- Define public Manager type and constructor New(cfg Config) (*Manager, error).
- Implement minimal public methods: Alloc(size int) (Handle, error), Free(h Handle) error, Read(h Handle, off, n int) ([]byte, error), Write(h Handle, off int, data []byte) error.
- Keep Handle opaque and stable; delegate placement and translation to internal/vmm and internal/store.

2. memory/types.go
- Define public Handle type (opaque virtual allocation ID).
- Define Config with DiskSize and PlacementStrategy (optional in v1; defaults to first-fit).
- Define minimal Stats struct only if needed by tests.

3. memory/errors.go
- Define sentinel public errors: ErrOutOfMemory, ErrInvalidHandle, ErrDoubleFree, ErrInvalidSize, ErrOutOfBounds.
- Keep error semantics documented for exact behavioral tests.

4. memory/manager_test.go
- Table-driven API tests for virtual allocations.
- Cases: fragmented allocation success via multiple segments, read/write across boundaries, invalid handle, double free, bounds checks.

5. internal/store/disk.go
- Define Disk struct wrapping []byte simulated disk.
- Provide internal low-level read/write by physical offset with strict bounds checks.

6. internal/store/free_list.go
- Define free extent representation (offset, size) and free-list manager.
- Implement first-fit extent reservation, split, and coalesce.
- Expose Reserve(size int) ([]Extent, error) that may return multiple extents.

7. internal/vmm/table.go
- Define mapping table: handle ID -> ordered list of segments (logical start, physical offset, length).
- Provide insert/remove/lookup operations and integrity checks.

8. internal/vmm/translate.go
- Implement logical-to-physical translation helpers used by Read/Write.
- Map (handle, logical offset, length) to one or more physical spans.
- Validate bounds and return deterministic translation errors.

9. internal/vmm/vmm_test.go
- Tests for mapping table and translation logic: multi-segment traversal, boundary edges, invalid offsets.

10. internal/store/store_test.go
- Tests for free-list/extents behavior: split/coalesce correctness and fragmentation scenarios.
- Verify Reserve can satisfy requests via multiple extents when contiguous space is unavailable.



**Alloc(size) Flow (sequence + layer ownership)**
1. Public API entry
- Layer: memory/manager.
- Validate input size (> 0), normalize request, and return ErrInvalidSize on bad input.

2. Reserve physical storage extents
- Layer: internal/store.
- Call Reserve(size) using first-fit.
- Result may include multiple extents when no single contiguous extent can satisfy the request.
- On failure return ErrOutOfMemory.

3. Build virtual segment mapping
- Layer: memory/manager orchestrating internal/vmm.
- Convert reserved extents into ordered segments spanning logical range [0, size).

4. Create handle + persist mapping
- Layer: internal/vmm.
- Create a new opaque handle ID and store handle -> ordered segments mapping.

5. Rollback contract on mapping failure
- Layer: memory/manager + internal/store.
- If VMM mapping insert fails after reserve succeeded, Manager must release reserved extents back to Store and return an error.

6. Return handle
- Layer: memory/manager.
- Return stable opaque handle to caller.

**Alloc layer responsibilities**
- memory/manager: validation, orchestration, rollback, and public error mapping.
- internal/store: physical extent reservation/splitting.
- internal/vmm: handle lifecycle and segment mapping persistence.

**Free(handle) Flow (sequence + layer ownership)**
1. Public API entry
- Layer: memory/manager.
- Validate handle shape and return ErrInvalidHandle for obviously invalid input.

2. Remove mapping first and retrieve released extents
- Layer: internal/vmm.
- Atomically remove handle -> segments mapping and return corresponding physical extents.
- If handle is missing, return ErrInvalidHandle (covers double-free behavior in v1).

3. Release extents back to free space
- Layer: internal/store.
- Manager calls Store.Release(extents).
- Store owns insertion of returned extents into the free list.

4. Coalesce adjacent free extents (internal)
- Layer: internal/store.
- Coalescing is performed internally as part of Release.
- Coalesce is not exposed as a public Store API.

5. Return result
- Layer: memory/manager.
- Return success on complete release/coalesce.
- Return error if Store reports an invariant/release failure.

**Free layer responsibilities**
- memory/manager: orchestration and public error mapping.
- internal/vmm: mapping lifecycle (remove + return mapped extents).
- internal/store: physical release and coalescing.



**Read(handle, off, n) Flow (sequence + layer ownership)**
1. Public API entry and argument validation
- Layer: memory/manager.
- Validate handle token shape, off >= 0, n >= 0, and fast-path n == 0 (return empty result).
- Return ErrInvalidHandle / ErrOutOfBounds / ErrInvalidSize as appropriate.

2. Resolve logical range against mapping
- Layer: internal/vmm.
- Lookup handle mapping and validate requested logical range [off, off+n).
- Translate logical range to ordered physical spans (one or many segments).
- Return ErrInvalidHandle if handle not found; ErrOutOfBounds if range exceeds allocation size.

3. Read bytes from physical store
- Layer: internal/store.
- Manager calls concrete internal Store read helpers (for example ReadAt(offset, n)); v1 does not introduce a Store interface.
- For each translated span, Store performs bounded reads by physical offset and length from backing []byte.
- Append/copy data into output buffer in logical order.

4. Return assembled result
- Layer: memory/manager.
- Return concatenated byte slice with length n.
- Return deterministic translated/store error if any span read fails.

**Read layer responsibilities**
- memory/manager: API validation, orchestration, output assembly, public error mapping.
- internal/vmm: handle lookup and logical-to-physical translation.
- internal/store: bounded physical reads via concrete internal methods; not exposed as public API and not abstracted behind a Store interface in v1.


**Write(handle, off, data) Flow (sequence + layer ownership)**
1. Public API entry and argument validation
- Layer: memory/manager.
- Validate handle token shape, off >= 0, and len(data) >= 0.
- Fast-path len(data) == 0 returns success.
- Return ErrInvalidHandle / ErrOutOfBounds / ErrInvalidSize as appropriate.

2. Resolve logical write range against mapping
- Layer: internal/vmm.
- Lookup handle mapping and validate requested logical range [off, off+len(data)).
- Translate logical write range to ordered physical spans (one or many segments).
- Return ErrInvalidHandle if handle not found; ErrOutOfBounds if range exceeds allocation size.

3. Write bytes to physical store
- Layer: internal/store.
- Manager calls concrete internal Store write helpers (for example WriteAt(offset, chunk)); v1 does not introduce a Store interface.
- For each translated span, Store performs bounded writes by physical offset from backing []byte.

4. Return result
- Layer: memory/manager.
- Return success if all span writes complete.
- Return deterministic translated/store error if any span write fails.

**Write layer responsibilities**
- memory/manager: API validation, orchestration, chunk slicing across spans, public error mapping.
- internal/vmm: handle lookup and logical-to-physical translation for write ranges.
- internal/store: bounded physical writes via concrete internal methods; not exposed as public API and not abstracted behind a Store interface in v1.