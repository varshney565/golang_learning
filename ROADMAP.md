# Go Learning Roadmap

Run any topic: `go run ./01_types/` or `go run ./09_goroutines/`

Read them in order — each topic builds on the previous ones.

---

## Basics (01 – 05)
| Directory | Topics |
|-----------|--------|
| `01_types/` | Zero values, type definitions, type assertions, type switches, iota |
| `02_pointers/` | Pointer basics, pass by reference, nil pointers, pointer receivers |
| `03_slices/` | Internals (len/cap), append reallocation, copy, shared array gotcha |
| `04_maps/` | CRUD, nil map gotcha, comma-ok, delete, sets, grouping |
| `05_structs/` | Embedding, anonymous fields, struct tags, comparison |

## Functions & Interfaces (06 – 08)
| Directory | Topics |
|-----------|--------|
| `06_functions/` | Multiple returns, variadic, function types, closures, loop capture bug |
| `07_interfaces/` | Implicit impl, composition, Stringer, error, empty interface, nil gotcha |
| `08_defer_panic_recover/` | Defer LIFO order, named returns, panic, recover patterns |

## Concurrency (09 – 16)
| Directory | Topics |
|-----------|--------|
| `09_goroutines/` | Launch, WaitGroup, goroutine leaks, data races, GOMAXPROCS |
| `10_channels/` | Unbuffered, buffered, direction, closing, pipeline, semaphore |
| `11_select/` | Basics, default, timeout pattern, fan-in, nil channel trick |
| `12_sync/` | Mutex, RWMutex, Once, atomic, sync.Map, sync.Pool |
| `13_concurrency_patterns/` | Worker pool, fan-out/in, pipeline+cancel, rate limiter, errgroup |
| `14_context/` | WithCancel, WithTimeout, WithDeadline, WithValue, propagation |
| `15_sync_cond/` | sync.Cond: Signal, Broadcast, bounded buffer, cond vs channel |
| `16_deadlocks/` | Channel deadlock, mutex deadlock, circular wait, goroutine leaks |

## Error Handling (17)
| Directory | Topics |
|-----------|--------|
| `17_errors/` | Sentinel errors, custom types, %w wrapping, errors.Is/As, multierror |

## Standard Library (18 – 24)
| Directory | Topics |
|-----------|--------|
| `18_io/` | io.Reader/Writer, io.Copy, bufio, os file operations |
| `19_http/` | Handlers, ServeMux, middleware, HTTP client, httptest |
| `20_json/` | Marshal/unmarshal, tags, custom marshaler, streaming, RawMessage |
| `21_time/` | Timer, Ticker, time.After leak trap, stop correctly, formatting |
| `22_sort_strings_strconv/` | sort.Slice, binary search, strings package, strconv conversions |
| `23_modules/` | go.mod, module layout, init(), build tags, go:generate |
| `24_os_exec/` | Run commands, pipes, stdin, context kill, env vars, injection safety |

## Testing & Tooling (25 – 26)
| Directory | Topics |
|-----------|--------|
| `25_testing/` | Table-driven tests, subtests, mocks, httptest, benchmarks, TestMain |
| `26_pprof/` | CPU/heap/goroutine profiles, HTTP pprof, flame graphs, workflow |

## Advanced / Interview Level (27 – 36)
| Directory | Topics |
|-----------|--------|
| `27_generics/` | Type params, constraints, Map/Filter/Reduce, generic data structures |
| `28_reflection/` | TypeOf/ValueOf, struct inspection, dynamic calls, custom validator |
| `29_memory_gc/` | Stack vs heap, escape analysis, GC internals, GOGC, struct padding |
| `30_design_patterns/` | Functional options, builder, observer, strategy, decorator |
| `31_database_sql/` | Connection pool, QueryRow, transactions, prepared stmts, repo pattern |
| `32_graceful_shutdown/` | SIGINT/SIGTERM, http.Server.Shutdown, WaitGroup drain, context cancel |
| `33_http_roundtripper/` | Custom transport, logging/auth/retry middleware, chain pattern |
| `34_embed/` | //go:embed, embed.FS, serving static files, migrations |
| `35_slog/` | log/slog, levels, JSON/text handlers, slog.With, custom handler |
| `36_interview_concepts/` | GMP scheduler, string internals, interface internals, all gotchas |
