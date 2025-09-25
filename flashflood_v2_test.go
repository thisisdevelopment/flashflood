package flashflood_test

import (
	"fmt"
	"testing"
	"time"

	flashflood "github.com/thisisdevelopment/flashflood/v2"
)

// TestBasicGenericUsage tests basic functionality with different types
func TestBasicGenericUsage(t *testing.T) {
	// Test with strings
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      100 * time.Millisecond,
	})

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("Error getting channel: %v", err)
	}

	ff.Push("item1", "item2", "item3", "item4")

	// Should receive items without type casting
	for i := 0; i < 4; i++ {
		select {
		case item := <-ch:
			// item is guaranteed to be string - no casting needed!
			if len(item) == 0 {
				t.Errorf("Received empty string")
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout waiting for item %d", i)
		}
	}

	ff.Close()
}

// TestGenericIntUsage tests with integer type
func TestGenericIntUsage(t *testing.T) {
	ff := flashflood.New[int](&flashflood.Opts{
		BufferAmount: 2,
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	ff.Push(1, 2, 3)

	for i := 0; i < 3; i++ {
		select {
		case item := <-ch:
			// item is guaranteed to be int
			if item < 1 || item > 3 {
				t.Errorf("Unexpected item value: %d", item)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout waiting for item %d", i)
		}
	}

	ff.Close()
}

// TestChunkedOutput tests the chunked output functionality
func TestChunkedOutput(t *testing.T) {
	// Create FlashFlood that outputs chunks of strings
	ff := flashflood.New[[]string](&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3,
		Timeout:      100 * time.Millisecond,
	})

	// Add the chunking transformation
	ff.AddFunc(flashflood.FuncMergeChunkedElements[string]())

	ch, _ := ff.GetChan()

	// Push individual string slices (each wrapped for chunking)
	ff.Push([]string{"item1"}, []string{"item2"}, []string{"item3"}, []string{"item4"}, []string{"item5"})

	// First chunk should have 3 items
	select {
	case chunk := <-ch:
		// chunk is []string - no casting needed!
		if len(chunk) != 3 {
			t.Errorf("Expected chunk of 3 items, got %d", len(chunk))
		}
		if chunk[0] != "item1" || chunk[1] != "item2" || chunk[2] != "item3" {
			t.Errorf("Unexpected chunk contents: %v", chunk)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for first chunk")
	}

	// Wait for timeout to flush remaining items
	time.Sleep(150 * time.Millisecond)

	// Second chunk should have 2 items (timeout flush)
	select {
	case chunk := <-ch:
		if len(chunk) != 2 {
			t.Errorf("Expected chunk of 2 items, got %d", len(chunk))
		}
		if chunk[0] != "item4" || chunk[1] != "item5" {
			t.Errorf("Unexpected chunk contents: %v", chunk)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for second chunk")
	}

	ff.Close()
}

// TestCustomStruct tests with custom struct type
func TestCustomStruct(t *testing.T) {
	type Record struct {
		ID   int
		Data string
	}

	ff := flashflood.New[Record](&flashflood.Opts{
		BufferAmount: 2,
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	ff.Push(Record{ID: 1, Data: "test1"}, Record{ID: 2, Data: "test2"})

	for i := 0; i < 2; i++ {
		select {
		case record := <-ch:
			// record is guaranteed to be Record type
			if record.ID != i+1 {
				t.Errorf("Expected ID %d, got %d", i+1, record.ID)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout waiting for record %d", i)
		}
	}

	ff.Close()
}

// TestByteSliceProcessing tests byte slice processing without special functions
func TestByteSliceProcessing(t *testing.T) {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 2,
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	ff.Push([]byte{1, 2, 3}, []byte{4, 5, 6})

	for i := 0; i < 2; i++ {
		select {
		case byteSlice := <-ch:
			// byteSlice is guaranteed to be []byte
			if len(byteSlice) != 3 {
				t.Errorf("Expected byte slice of length 3, got %d", len(byteSlice))
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("Timeout waiting for byte slice %d", i)
		}
	}

	ff.Close()
}

// TestGateAmount tests gate mechanism with generics
func TestGateAmount(t *testing.T) {
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 2, // Small buffer so items overflow
		GateAmount:   3, // Release exactly 3 items at a time
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	// Push 5 items - buffer=2 so 3 will overflow and trigger gate mechanism
	ff.Push("a", "b", "c", "d", "e")

	// Collect all items
	received := 0
	items := make([]string, 0)

	// Wait for immediate overflow drain
	timeout := time.After(50 * time.Millisecond)
drainLoop:
	for {
		select {
		case item := <-ch:
			received++
			items = append(items, item)
		case <-timeout:
			break drainLoop
		}
	}

	// Wait for timeout to flush remaining items
	time.Sleep(150 * time.Millisecond)

	// Collect remaining items
	for {
		select {
		case item := <-ch:
			received++
			items = append(items, item)
		default:
			goto done
		}
	}
done:

	if received != 5 {
		t.Errorf("Expected 5 total items, got %d: %v", received, items)
	}

	ff.Close()
}

// Example functions for godoc

// ExampleNew demonstrates basic generic usage
func ExampleNew() {
	// Create a string buffer
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      100 * time.Millisecond,
	})

	ch, _ := ff.GetChan()
	ff.Push("item1", "item2", "item3", "item4")

	// Receive items - no type casting needed!
	for i := 0; i < 4; i++ {
		item := <-ch
		fmt.Printf("Received: %s\n", item)
	}

	ff.Close()
	// Output: Received: item1
	// Received: item2
	// Received: item3
	// Received: item4
}

// ExampleFuncMergeChunkedElements demonstrates chunked output
func ExampleFuncMergeChunkedElements() {
	// Create buffer that outputs chunks of strings
	ff := flashflood.New[[]string](&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   3,
		Timeout:      100 * time.Millisecond,
	})

	// FuncMergeChunkedElements is deprecated in v2
	ch, _ := ff.GetChan()

	ff.Push([]string{"a", "b", "c"}, []string{"d", "e"})

	// Wait for timeout to ensure all items are processed
	time.Sleep(150 * time.Millisecond)

	for {
		select {
		case chunk := <-ch:
			fmt.Printf("Chunk: %v\n", chunk)
		default:
			goto done
		}
	}
done:

	ff.Close()
	// Output: Chunk: [a b c]
	// Chunk: [d e]
}
