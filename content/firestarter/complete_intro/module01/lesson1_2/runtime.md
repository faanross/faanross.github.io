---
showTableOfContents: true
title: "Part 4 - Go Runtime Internals Relevant to Evasion"
type: "page"
---

## **PART 4: GO RUNTIME INTERNALS RELEVANT TO EVASION**

### **Understanding What's Under the Hood**

To build evasive tools in Go, you must understand the **Go runtime** and how it affects your implants' behavior and detectability.

```
┌──────────────────────────────────────────────────────────────┐
│                    GO RUNTIME COMPONENTS                     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  RUNTIME COMPONENTS (Included in Every Binary):              │
│                                                              │
│  1. MEMORY ALLOCATOR                                         │
│     • Manages heap (new/make)                                │
│     • mheap, mcentral, mcache structures                     │
│     • Creates distinct memory patterns                       │
│                                                              │
│  2. GARBAGE COLLECTOR                                        │
│     • Mark-and-sweep algorithm                               │
│     • Concurrent collection (since Go 1.5)                   │
│     • Periodic scans (behavioral IOC)                        │
│                                                              │
│  3. GOROUTINE SCHEDULER                                      │
│     • M:N scheduler (M goroutines on N OS threads)           │
│     • Work-stealing algorithm                                │
│     • Creates unique thread patterns                         │
│                                                              │
│  4. TYPE SYSTEM                                              │
│     • Reflection metadata                                    │
│     • Interface tables                                       │
│     • Type descriptors in binary                             │
│                                                              │
│  5. PANIC/RECOVER MECHANISM                                  │
│     • Stack unwinding                                        │
│     • Defer handling                                         │
│     • Error stack traces                                     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### **Runtime Evasion Considerations**

**1. Heap Behaviour and Memory Patterns**

```go
// Go's heap allocation is detectable by pattern

// Standard allocation (leaves signature)
data := make([]byte, 1024*1024) // 1MB allocation
// Runtime manages this, creates patterns

// Behavioral Analysis Can Detect:
// • Allocation sizes (powers of 2 common in Go)
// • GC timing
// • Heap growth patterns

// Evasion Technique: Custom allocators
import (
    "syscall"
    "unsafe"
)

// Allocate via Windows API (bypasses Go heap)
func stealthAlloc(size uintptr) unsafe.Pointer {
    addr, _, _ := syscall.Syscall6(
        procVirtualAlloc.Addr(),
        4,
        0,
        size,
        syscall.MEM_COMMIT|syscall.MEM_RESERVE,
        syscall.PAGE_READWRITE,
        0,
        0,
    )
    return unsafe.Pointer(addr)
}

// Use case: shellcode storage (avoid Go heap)
shellcodePtr := stealthAlloc(uintptr(len(shellcode)))
// Copy shellcode to non-Go memory
```

**2. Garbage Collector Manipulation**

```go
package main

import (
    "runtime"
    "runtime/debug"
    "time"
)

func main() {
    // Understanding GC parameters
    
    // 1. Check current GC stats
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    println("GC cycles:", stats.NumGC)
    println("Pause time:", stats.PauseTotalNs)
    
    // 2. Control GC behavior
    
    // Disable GC during sensitive operations
    debug.SetGCPercent(-1)
    
    // Perform sensitive work (process injection, etc.)
    injectPayload()
    
    // Re-enable with custom target
    debug.SetGCPercent(200) // Less frequent collections
    
    // 3. Force GC at specific times (hide timing patterns)
    for {
        doWork()
        
        // Random GC timing (defeats periodic detection)
        if time.Now().Unix() % 17 == 0 {
            runtime.GC()
        }
        
        time.Sleep(randomDuration())
    }
}

// 4. Minimize allocations in hot paths
func criticalOperation() {
    // Bad: allocates on every call
    buffer := make([]byte, 4096)
    
    // Good: reuse buffers
    // (use sync.Pool in production)
}
```

**3. Goroutine Scheduler Implications**

```go
// Go creates OS threads for goroutines
// This can be detected via process analysis

import (
    "runtime"
)

func init() {
    // Control number of OS threads
    // Default: GOMAXPROCS = number of CPU cores
    
    // Reduce thread count (less visible)
    runtime.GOMAXPROCS(1)
    
    // Or match normal application behavior
    // (e.g., if masquerading as 4-thread app)
    runtime.GOMAXPROCS(4)
}

// Goroutines are lightweight but still detectable
func stealthyOperation() {
    // Obvious: many goroutines
    for i := 0; i < 1000; i++ {
        go doTask() // 1000 goroutines created
    }
    
    // Stealthier: sequential or limited concurrency
    maxWorkers := 4
    sem := make(chan struct{}, maxWorkers)
    
    for i := 0; i < 1000; i++ {
        sem <- struct{}{}
        go func() {
            defer func() { <-sem }()
            doTask()
        }()
    }
}
```

**4. Stack Traces and Debugging Information**

```go
// Go binaries contain debugging info by default

// Default build: ~6MB, includes symbols
go build implant.go

// Stripped build: ~3MB, minimal info
go build -ldflags="-s -w" implant.go
// -s: strip symbol table
// -w: strip DWARF debugging

// Even more minimal
go build -ldflags="-s -w" -trimpath implant.go
// -trimpath: remove file system paths

// Analysis of what's removed:
// WITH symbols:    Function names, line numbers, file paths
// WITHOUT symbols: Harder to reverse engineer, fewer IOCs
```

**Comparison of Build Outputs:**

```bash
# Build with debug info
$ go build implant.go
$ ls -lh implant.exe
-rwxr-xr-x 1 user user 6.1M implant.exe

$ strings implant.exe | grep main.
main.connectC2
main.executeCommand
main.main
runtime.main
# Exposes function names!

# Build stripped
$ go build -ldflags="-s -w" -trimpath implant.go
$ ls -lh implant.exe  
-rwxr-xr-x 1 user user 3.2M implant.exe

$ strings implant.exe | grep main.
# Much less information exposed
```

### **Runtime Detection Vectors**

**What Defenders Can Detect:**

```
STATIC ANALYSIS:
✓ Go build ID in binary
✓ Go version string
✓ Runtime function names (even stripped)
✓ Standard library patterns
✓ Type information (partially)

BEHAVIOURAL ANALYSIS:
✓ GC pause patterns
✓ Goroutine scheduling behavior  
✓ Heap allocation patterns
✓ Thread creation patterns
✓ Memory layout (Go-specific)

NETWORK ANALYSIS:
✓ HTTP library fingerprints
✓ TLS handshake patterns (crypto/tls)
✓ DNS query patterns (net package)
```

**Mitigation Strategies:**

```go
// 1. Remove Go version string
go build -ldflags="-s -w -X runtime.buildVersion=" implant.go

// 2. Use custom HTTP client (not net/http directly)
// Mimic legitimate applications

// 3. Encrypt all strings (including library calls)
// Use compile-time obfuscation

// 4. Custom syscalls (bypass Go's syscall package)
// Direct assembly or syscall stubs

// 5. Obfuscate with garble
garble -literals -tiny build implant.go
// -literals: encrypt string literals
// -tiny: optimize for size
```

### **Go Runtime: Friend or Foe?**

```
FRIEND (Advantages):
✓ Stability: GC prevents memory leaks
✓ Safety: Bounds checking prevents crashes
✓ Concurrency: Goroutines simplify threading
✓ Portability: Runtime abstracts OS differences

FOE (Disadvantages):
✗ Fingerprints: Runtime behavior detectable
✗ Size: Runtime adds MB to binary
✗ Predictability: GC/scheduler patterns
✗ Metadata: Type info aids analysis

VERDICT:
For most operations, runtime benefits outweigh costs.
For maximum stealth, consider C/Rust or hybrid approach.
```

---



[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./comparison.md" >}})
[|NEXT|]({{< ref "./setup.md" >}})