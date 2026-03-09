# Go Learning Roadmap

Run any file: `go run ./phase1/01_types/`

## Phase 1 — Solidify the Basics
| File | Topics |
|------|--------|
| `phase1/01_types/` | Zero values, type definitions, type assertions, type switches, iota |
| `phase1/02_pointers/` | Pointer basics, pass by ref, nil pointers, pointer receivers |
| `phase1/03_slices/` | Internals (len/cap), append reallocation, copy, 2D slices |
| `phase1/04_maps/` | CRUD, nil map gotcha, comma-ok, delete, sets, grouping |
| `phase1/05_structs/` | Embedding, anonymous fields, struct tags, comparison |

## Phase 2 — Functions & Interfaces
| File | Topics |
|------|--------|
| `phase2/01_functions/` | Multiple returns, variadic, function types, closures, IIFE |
| `phase2/02_interfaces/` | Implicit impl, composition, Stringer, error, empty interface, nil gotcha |
| `phase2/03_defer_panic_recover/` | Defer order, LIFO, closure vs arg, panic, recover patterns |

## Phase 3 — Concurrency
| File | Topics |
|------|--------|
| `phase3/01_goroutines/` | Launch, WaitGroup, leaks, data races, GOMAXPROCS |
| `phase3/02_channels/` | Unbuffered, buffered, direction, close, pipeline, semaphore |
| `phase3/03_select/` | Basics, default, timeout, fan-in, nil channel trick |
| `phase3/04_sync/` | Mutex, RWMutex, Once, atomic, sync.Map, sync.Pool |
| `phase3/05_patterns/` | Worker pool, fan-out/in, pipeline+cancel, rate limiter, errgroup |
| `phase3/06_context/` | WithCancel, WithTimeout, WithDeadline, WithValue, propagation |

## Phase 4 — Error Handling
| File | Topics |
|------|--------|
| `phase4/01_error_handling/` | Sentinel errors, custom types, %w wrapping, Is/As, multierror |

## Phase 5 — Packages & Modules
| File | Topics |
|------|--------|
| `phase5/01_modules/` | go.mod, module layout, init(), build tags, go:generate |

## Phase 6 — Standard Library
| File | Topics |
|------|--------|
| `phase6/01_io/` | io.Reader/Writer, io.Copy, bufio, os file ops |
| `phase6/02_http/` | Handlers, ServeMux, middleware, client, httptest |
| `phase6/03_json/` | Marshal/unmarshal, tags, custom marshaler, streaming, RawMessage |

## Phase 7 — Testing
| File | Topics |
|------|--------|
| `phase7/01_testing/` | Table-driven, subtests, mocks, httptest, benchmarks, TestMain |

## Advanced (Interview Level)
| File | Topics |
|------|--------|
| `advanced/01_generics/` | Type params, constraints, Map/Filter/Reduce, generic data structures |
| `advanced/02_reflection/` | TypeOf/ValueOf, struct inspection, dynamic calls, custom validator |
| `advanced/03_memory_gc/` | Stack vs heap, escape analysis, GC internals, GOGC, struct padding |
| `advanced/04_design_patterns/` | Functional options, builder, observer, strategy, decorator |
| `advanced/05_interview_concepts/` | GMP scheduler, string internals, interface internals, gotchas |
