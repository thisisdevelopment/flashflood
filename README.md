- [Why FlashFlood?](#why-flashflood)
- [Key Features](#key-features)
- [Real-World Use Cases](#real-world-use-cases)
- [Quick Start](#quick-start)
- [Gate Mechanism - Predictable Batching](#gate-mechanism---predictable-batching)
- [Advanced Usage](#advanced-usage)
- [Performance](#performance)
- [API Reference](#api-reference)
- [About Us Th\[is\]](#about-us-this)
- [Contributing](#contributing)
- [License](#license)

# FlashFlood: High-Performance Ring Buffer

[![go report card](https://goreportcard.com/badge/github.com/thisisdevelopment/flashflood "go report card")](https://goreportcard.com/report/github.com/thisisdevelopment/flashflood)
[![codecov](https://codecov.io/gh/thisisdevelopment/flashflood/branch/master/graph/badge.svg)](https://codecov.io/gh/thisisdevelopment/flashflood)
[![CircleCI](https://circleci.com/gh/thisisdevelopment/flashflood.svg?style=svg)](https://circleci.com/gh/thisisdevelopment/flashflood)
[![GoDoc](https://godoc.org/github.com/thisisdevelopment/flashflood?status.svg)](https://godoc.org/github.com/thisisdevelopment/flashflood)

FlashFlood is a **high-performance ring buffer** with advanced batching capabilities that goes far beyond what standard Go channels offer. While channels are excellent for simple producer-consumer scenarios, FlashFlood excels when you need **predictable batch sizes**, **automatic timeouts**, and **element transformations**.

## Why FlashFlood?

**Standard Go channels** force you to choose:
- Process elements one-by-one (inefficient for bulk operations)
- Build complex batching logic yourself (error-prone and verbose)
- Handle timeouts manually (more boilerplate)

**FlashFlood solves this** by providing:
- ✅ **Guaranteed batch sizes** via the gate mechanism
- ✅ **Automatic timeout handling** for incomplete batches
- ✅ **Element transformations** with FuncStack callbacks
- ✅ **Thread-safe operations** with minimal overhead
- ✅ **Back-pressure control** through configurable buffer sizes

## Key Features

### 🎯 **Gate-Based Batching**
Set `GateAmount: 10` and receive exactly 10 elements at a time - perfect for database bulk inserts, API batch calls, or file writing operations.

### ⏱️ **Smart Timeout Handling**
Buffer doesn't fill to gate size? No problem - automatic timeout ensures data still flows even during low-traffic periods.

### 🔄 **Element Transformations**
Apply transformations to batches before output using FuncStack callbacks - merge byte arrays, aggregate data, or format for APIs.

### 🚀 **High Performance**
Benchmarks show consistent performance with minimal allocations, even under heavy load.

### 🔒 **Thread-Safe**
Concurrent producers and consumers work seamlessly with internal mutex protection.

## Real-World Use Cases

### 📊 **Database Batch Inserts**
```go
// Collect exactly 100 records, then bulk insert
ff := flashflood.New(&flashflood.Opts{
    BufferAmount: 1000,
    GateAmount:   100,    // Always insert 100 records at once
    Timeout:      5*time.Second,  // Flush incomplete batches after 5s
})
```

### 🌐 **API Rate Limiting & Batching**
```go
// Group API calls into batches of 25 to stay under rate limits
ff := flashflood.New(&flashflood.Opts{
    GateAmount:   25,     // Batch 25 API calls together
    Timeout:      2*time.Second,  // Don't wait longer than 2s
})
```

### 📝 **Log Aggregation**
```go
// Collect log entries and write to disk efficiently
ff := flashflood.New(&flashflood.Opts{
    GateAmount:   50,     // Write 50 log entries at once
    Timeout:      1*time.Second,  // Flush every second for real-time monitoring
})
```

### 📦 **Message Queue Publishing**
```go
// Batch messages for better throughput
ff := flashflood.New(&flashflood.Opts{
    GateAmount:   20,     // Publish 20 messages per batch
    Timeout:      500*time.Millisecond,
})
```

## Quick Start

### Basic Usage
```go
package main

import (
    "fmt"
    "time"
    "github.com/thisisdevelopment/flashflood"
)

func main() {
    // Create buffer that flushes when 3+ elements or after 250ms
    ff := flashflood.New(&flashflood.Opts{
        BufferAmount: 10,  // Internal buffer size
        Timeout: 250 * time.Millisecond,
    })

    // Get the overflow channel
    ch, _ := ff.GetChan()

    // Add elements to buffer
    ff.Push("item1", "item2", "item3", "item4")

    // Receive flushed elements
    for i := 0; i < 4; i++ {
        select {
        case item := <-ch:
            fmt.Printf("Received: %v\n", item)
        }
    }
}
```

## Gate Mechanism - Predictable Batching

**The gate mechanism is FlashFlood's killer feature** - it guarantees consistent batch sizes that your receivers can count on.

### How Gates Work
```go
ff := flashflood.New(&flashflood.Opts{
    BufferAmount: 100,    // Internal buffer holds 100 items
    GateAmount:   10,     // Release exactly 10 items at a time
    Timeout:      1*time.Second,  // Fallback: flush after 1 second
})
```

**Behavior:**
- Buffer collects elements until it has **exactly `GateAmount`** items
- Then releases **all `GateAmount` items at once**
- If timeout occurs before gate fills, flushes whatever is available
- Receiver always knows: "I'll get exactly 10 items, or it's a timeout flush"

### Gate Examples

#### Database Batch Inserts
```go
// Always insert exactly 50 records at once
ff := flashflood.New(&flashflood.Opts{
    GateAmount: 50,
    Timeout:    5*time.Second,
})

// Add function to keep batched elements grouped together
ff.AddFunc(ff.FuncMergeChunkedElements())

// Your receiver gets batches: either 50 records (normal) or <50 (timeout)
for {
    select {
    case recordBatch := <-ch:
        records := recordBatch.([]interface{})
        if len(records) == 50 {
            // Normal batch - optimal performance
            db.BulkInsert(records)
        } else {
            // Timeout batch - still insert but log it
            log.Printf("Timeout flush: %d records", len(records))
            db.BulkInsert(records)
        }
    }
}
```

#### Byte Stream Processing
```go
// Process data in 1KB chunks
ff := flashflood.New(&flashflood.Opts{
    GateAmount: 1024,  // Exactly 1KB chunks
    Timeout:    100*time.Millisecond,
})

ff.AddFunc(ff.FuncMergeBytes())  // Merge individual bytes into single chunk

// Receiver gets exactly 1KB chunks for optimal processing
for {
    select {
    case chunk := <-ch:
        // chunk is always []byte of exactly 1024 bytes (or timeout)
        processChunk(chunk.([]byte))
    }
}
```

## Advanced Usage

### Element Transformations with FuncStack
Apply functions to batches before they're sent to the channel:

```go
ff := flashflood.New(&flashflood.Opts{
    GateAmount: 5,
})

// Add custom transformation
ff.AddFunc(func(items []interface{}, ff *flashflood.FlashFlood) []interface{} {
    // Transform each item (e.g., add timestamp)
    result := make([]interface{}, len(items))
    for i, item := range items {
        result[i] = fmt.Sprintf("%v_processed_at_%d", item, time.Now().Unix())
    }
    return result
})
```

### Built-in Transformation Functions
```go
// Merge byte slices into single byte array
ff.AddFunc(ff.FuncMergeBytes())

// Return individual bytes from byte slices
ff.AddFunc(ff.FuncReturnIndividualBytes())

// Keep elements grouped in chunks
ff.AddFunc(ff.FuncMergeChunkedElements())
```

### Multiple Transformations
```go
// Chain multiple transformations - they execute in order
ff.AddFunc(transformFunc1)  // Executes first
ff.AddFunc(transformFunc2)  // Then this
ff.AddFunc(transformFunc3)  // Finally this
```

### Manual Control
```go
// Force flush current buffer to channel
ff.Drain(true, false)  // (toChannel=true, respectGate=false)

// Get elements directly without using channel
items, _ := ff.Get(10)  // Get up to 10 items

// Check buffer status
count := ff.Count()  // How many items in buffer

// Clear buffer
ff.Purge()

// Update activity (resets timeout)
ff.Ping()
```

## Performance

FlashFlood delivers consistent high performance across different scenarios:

```
Operation                                    Ops/sec    ns/op    Allocs
─────────────────────────────────────────────────────────────────────
BenchmarkPushChan-16                        4.2M       274      1
BenchmarkPushNoChan-16                       6.1M       225      1
BenchmarkPushChanGate-16                     5.2M       239      1
BenchmarkPushChanBigBuffer-16                5.8M       209      1
BenchmarkWithGet-16                          4.2M       299      2

With Callback Functions:
BenchmarkPushChanWithCallback/gate-1-16     3.6M       327      1
BenchmarkPushChanWithCallback/gate-16-16    3.3M       355      1
BenchmarkPushChanWithCallback/gate-1024-16  3.3M       357      1
```

**Key Performance Benefits:**
- **Minimal allocations**: Typically just 1 allocation per operation
- **Consistent latency**: Performance stays stable regardless of gate size
- **High throughput**: 3-6 million operations per second
- **Low overhead**: Callback functions add minimal performance cost

## API Reference

### Core Methods

```go
// Create new FlashFlood instance
ff := flashflood.New(&flashflood.Opts{
    BufferAmount:  100,               // Internal buffer size
    GateAmount:    10,                // Batch size for releases
    Timeout:       1*time.Second,     // Auto-flush timeout
    TickerTime:    10*time.Millisecond, // Timeout check frequency
    ChannelBuffer: 1000,              // Output channel buffer size
    Debug:         false,             // Enable debug output
})

// Get output channel
ch, err := ff.GetChan()

// Add elements
ff.Push(item1, item2, item3)
ff.Unshift(priority_item)  // Add to front of buffer

// Manual operations
ff.Drain(true, false)      // Force flush to channel (toChannel, respectGate)
items, _ := ff.Get(5)      // Get up to 5 items directly (returns []interface{})
ff.GetOnChan(5)           // Get 5 items and send to channel
ff.Purge()                // Clear buffer (returns error)
count := ff.Count()       // Buffer size (returns uint64)
ff.Ping()                 // Reset timeout (no return value)
ff.Close()                // Cleanup resources

// Add transformations
ff.AddFunc(myTransformFunction)
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `BufferAmount` | 256 | Internal buffer size before overflow |
| `GateAmount` | 1 | Number of elements to release at once |
| `Timeout` | 100ms | Time before auto-flushing incomplete batches |
| `TickerTime` | 10ms | How often to check for timeouts |
| `ChannelBuffer` | 4096 | Output channel buffer size |
| `FlushTimeout` | 0 | Alternative timeout for different flush behavior |
| `FlushEnabled` | false | Enable separate flush timeout logic |
| `Debug` | false | Print debug information |
| `DisableRingUntilChanActive` | false | Prevent overflow until channel is retrieved |

**Full documentation and more examples:** https://godoc.org/github.com/thisisdevelopment/flashflood

## About Us Th[is]

[This.nl](https://this.nl) is a digital agency based in Utrecht, the Netherlands, specializing in crafting high-performance, resilient, and scalable digital solutions, api's, microservices, and more. FlashFlood represents our commitment to building robust, efficient tooling that solves real-world performance challenges in Go applications.

# Contributing
You can help to deliver a better flashflood buffer, check out how you can do things [CONTRIBUTING.md](CONTRIBUTING.md)

# License 
© [This is Development BV](https://www.thisisdevelopment.nl), 2019~time.Now()
Released under the [MIT License](https://github.com/thisisdevelopment/flashflood/blob/master/LICENSE)
