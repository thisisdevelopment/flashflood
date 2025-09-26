package flashflood_test

import (
	"fmt"
	"time"

	flashflood "github.com/thisisdevelopment/flashflood/v2"
)

// Example demonstrating basic usage from README
func ExampleFlashFlood_basicUsage() {
	// Create buffer that flushes when buffer fills or after 250ms
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 3, // Internal buffer size
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
			// item is guaranteed to be string - no casting needed!
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

// Example showing individual item processing (vs batched)
func ExampleFlashFlood_individualProcessing() {
	// Process individual records
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 2, // Small buffer to trigger overflow
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	// Add individual records
	ff.Push("record1", "record2", "record3", "record4", "record5")

	// Process individual records
	count := 0
	timeout := time.After(200 * time.Millisecond)
	for count < 5 {
		select {
		case item := <-ch:
			// item is guaranteed to be string - no casting needed!
			fmt.Printf("Processing: %v\n", item)
			count++
		case <-timeout:
			// Handle any remaining items
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
	// Create buffer that outputs batches of strings
	ff := flashflood.New[[]string](&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3, // Always process 3 records at once
		Timeout:      1 * time.Second,
	})

	// Add function to keep batched elements grouped
	ff.AddFunc(flashflood.FuncMergeChunkedElements[string]())

	ch, _ := ff.GetChan()

	// Simulate adding database records individually for batching
	ff.Push([]string{"record1"}, []string{"record2"}, []string{"record3"}, []string{"record4"}, []string{"record5"})

	// Wait for timeout to ensure all items are processed
	time.Sleep(1500 * time.Millisecond)

	// Process batches
	for {
		select {
		case batch := <-ch:
			// batch is []string - no casting needed!
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
	// Process individual byte slices
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 2, // Small buffer to trigger overflow
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	// Add byte slices
	ff.Push([]byte{0x1, 0x2, 0x3}, []byte{0x4})

	// Process byte slices
	processed := 0
	for processed < 2 {
		select {
		case chunk := <-ch:
			// chunk is []byte - no casting needed!
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
	ff := flashflood.New[string](&flashflood.Opts{
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
		// item is guaranteed to be string - no casting needed!
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
