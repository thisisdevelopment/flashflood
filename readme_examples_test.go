package flashflood_test

import (
	"fmt"
	"time"

	"github.com/thisisdevelopment/flashflood"
)

// Example demonstrating basic usage from README
func ExampleFlashFlood_basicUsage() {
	// Create buffer that flushes when buffer fills or after 250ms
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,  // Internal buffer size
		Timeout:      250 * time.Millisecond,
	})

	// Get the overflow channel
	ch, _ := ff.GetChan()

	// Add elements to buffer
	ff.Push("item1", "item2", "item3", "item4")

	// Receive flushed elements (first 3 will flush when buffer fills)
	for i := 0; i < 4; i++ {
		select {
		case item := <-ch:
			fmt.Printf("Received: %v\n", item)
		case <-time.After(500 * time.Millisecond):
			// Allow timeout to flush remaining items
			for {
				select {
				case item := <-ch:
					fmt.Printf("Received: %v\n", item)
				default:
					goto done
				}
			}
		}
	}
done:

	ff.Close()
	// Output: Received: item1
	// Received: item2
	// Received: item3
	// Received: item4
}

// Example demonstrating database batch inserts scenario - ORIGINAL VERSION
func ExampleFlashFlood_databaseBatching_original() {
	// Collect exactly 3 records, then bulk insert
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3,    // Always process 3 records at once
		Timeout:      1 * time.Second,
	})

	ch, _ := ff.GetChan()

	// Simulate adding database records
	ff.Push("record1", "record2", "record3", "record4", "record5")

	// Process first batch (exactly 3 items)
	count := 0
	for count < 5 {
		select {
		case item := <-ch:
			fmt.Printf("Processing: %v\n", item)
			count++
		case <-time.After(1500 * time.Millisecond):
			// Timeout will flush remaining items
			for {
				select {
				case item := <-ch:
					fmt.Printf("Processing: %v\n", item)
					count++
				default:
					goto done2
				}
			}
		}
	}
done2:

	ff.Close()
	// Output: Processing: record1
	// Processing: record2
	// Processing: record3
	// Processing: record4
	// Processing: record5
}

// Example demonstrating database batch inserts scenario - CORRECTED VERSION
func ExampleFlashFlood_databaseBatching() {
	// Collect exactly 3 records, then bulk insert
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3,    // Always process 3 records at once
		Timeout:      1 * time.Second,
	})

	// Add function to keep batched elements grouped
	ff.AddFunc(ff.FuncMergeChunkedElements())

	ch, _ := ff.GetChan()

	// Simulate adding database records
	ff.Push("record1", "record2", "record3", "record4", "record5")

	// Wait for timeout to ensure all items are processed
	time.Sleep(1500 * time.Millisecond)

	// Process batches
	for {
		select {
		case batch := <-ch:
			fmt.Printf("Batch: %v\n", batch)
		default:
			goto done2
		}
	}
done2:

	ff.Close()
	// Output: Batch: [record1 record2 record3]
	// Batch: [record4 record5]
}

// Example demonstrating byte stream processing with gates
func ExampleFlashFlood_byteProcessing() {
	// Process data in chunks of exactly 3 bytes
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3,  // Process 3 bytes at once
		Timeout:      100 * time.Millisecond,
	})

	// Add merge function to combine byte slices
	ff.AddFunc(ff.FuncMergeBytes())

	ch, _ := ff.GetChan()

	// Add individual bytes
	ff.Push([]byte{0x1}, []byte{0x2}, []byte{0x3}, []byte{0x4})

	// Process merged byte chunks
	processed := 0
	for processed < 2 {
		select {
		case chunk := <-ch:
			fmt.Printf("Chunk: %v\n", chunk)
			processed++
		case <-time.After(200 * time.Millisecond):
			// Handle timeout case
			for {
				select {
				case chunk := <-ch:
					fmt.Printf("Chunk: %v\n", chunk)
					processed++
				default:
					goto done3
				}
			}
		}
	}
done3:

	ff.Close()
	// Output: Chunk: [1 2 3]
	// Chunk: [4]
}

// Example demonstrating manual control operations
func ExampleFlashFlood_manualControl() {
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 5,
		Timeout:      1 * time.Second,
	})

	ch, _ := ff.GetChan()

	// Add some items
	ff.Push("item1", "item2", "item3")

	// Check buffer count
	fmt.Printf("Buffer count: %d\n", ff.Count())

	// Get items directly (without channel)
	items, _ := ff.Get(2)
	fmt.Printf("Got items: %v\n", items)

	// Check buffer count again
	fmt.Printf("Buffer count after Get: %d\n", ff.Count())

	// Add more items
	ff.Push("item4", "item5")

	// Force drain to channel
	ff.Drain(true, false)

	// Receive drained items
	select {
	case item := <-ch:
		fmt.Printf("Drained: %v\n", item)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("No items in channel")
	}

	// Purge any remaining items
	ff.Purge()
	fmt.Printf("Buffer count after Purge: %d\n", ff.Count())

	ff.Close()
	// Output: Buffer count: 3
	// Got items: [item1 item2]
	// Buffer count after Get: 1
	// Drained: item3
	// Buffer count after Purge: 0
}