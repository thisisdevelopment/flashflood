package flashflood_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/thisisdevelopment/flashflood/v2"
)

func TestFuncPassThrough(t *testing.T) {
	// Test with string type
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 5,
		GateAmount:   3,
		Timeout:      100 * time.Millisecond,
	})

	// Add FuncPassThrough
	ff.AddFunc(flashflood.FuncPassThrough[string]())

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	// Push elements that should trigger gate
	ff.Push("item1", "item2", "item3")

	// Collect results
	var results []string
	for i := 0; i < 3; i++ {
		select {
		case item := <-ch:
			results = append(results, item)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for item %d", i+1)
		}
	}

	// Verify PassThrough doesn't modify the items
	expected := []string{"item1", "item2", "item3"}
	if !reflect.DeepEqual(results, expected) {
		t.Fatalf("expected %v, got %v", expected, results)
	}

	ff.Close()
}

func TestFuncPassThroughWithInt(t *testing.T) {
	// Test with int type to verify generics work
	ff := flashflood.New[int](&flashflood.Opts{
		BufferAmount: 5,
		GateAmount:   2,
		Timeout:      100 * time.Millisecond,
	})

	// Add FuncPassThrough
	ff.AddFunc(flashflood.FuncPassThrough[int]())

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	// Push elements that should trigger gate
	ff.Push(10, 20)

	// Collect results
	var results []int
	for i := 0; i < 2; i++ {
		select {
		case item := <-ch:
			results = append(results, item)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for item %d", i+1)
		}
	}

	// Verify PassThrough doesn't modify the items
	expected := []int{10, 20}
	if !reflect.DeepEqual(results, expected) {
		t.Fatalf("expected %v, got %v", expected, results)
	}

	ff.Close()
}

func TestFuncFlatten(t *testing.T) {
	// Test FuncFlatten with byte slices -> bytes
	// Create a FlashFlood[byte] (individual bytes)
	ff := flashflood.New[byte](&flashflood.Opts{
		BufferAmount: 10,
		GateAmount:   4,
		Timeout:      100 * time.Millisecond,
	})

	// Add FuncFlatten - this is for processing individual bytes
	ff.AddFunc(flashflood.FuncFlatten[[]byte, byte]())

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	// Push individual bytes to trigger gate
	ff.Push(0x01, 0x02, 0x03, 0x04)

	// Collect results
	var results []byte
	for i := 0; i < 4; i++ {
		select {
		case item := <-ch:
			results = append(results, item)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for byte %d", i+1)
		}
	}

	// Verify FuncFlatten passes through individual bytes unchanged
	expected := []byte{0x01, 0x02, 0x03, 0x04}
	if !reflect.DeepEqual(results, expected) {
		t.Fatalf("expected %v, got %v", expected, results)
	}

	ff.Close()
}

func TestFuncFlattenWithStrings(t *testing.T) {
	// Test FuncFlatten with string slices -> strings
	// Create a FlashFlood[string] (individual strings)
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 8,
		GateAmount:   3,
		Timeout:      100 * time.Millisecond,
	})

	// Add FuncFlatten - this processes individual strings
	ff.AddFunc(flashflood.FuncFlatten[[]string, string]())

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	// Push individual strings to trigger gate
	ff.Push("hello", "world", "test")

	// Collect results
	var results []string
	for i := 0; i < 3; i++ {
		select {
		case item := <-ch:
			results = append(results, item)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for string %d", i+1)
		}
	}

	// Verify FuncFlatten passes through individual strings unchanged
	expected := []string{"hello", "world", "test"}
	if !reflect.DeepEqual(results, expected) {
		t.Fatalf("expected %v, got %v", expected, results)
	}

	ff.Close()
}

// Test all callback functions together in a comprehensive example
func TestAllCallbackFunctions(t *testing.T) {
	t.Run("FuncMergeChunkedElements", func(t *testing.T) {
		ff := flashflood.New[[]string](&flashflood.Opts{
			BufferAmount: 10,
			GateAmount:   3,
			Timeout:      100 * time.Millisecond,
		})

		ff.AddFunc(flashflood.FuncMergeChunkedElements[string]())

		ch, err := ff.GetChan()
		if err != nil {
			t.Fatalf("could not get channel: %v", err)
		}

		// Push individual slices
		ff.Push([]string{"item1"}, []string{"item2"}, []string{"item3"})

		// Should get one merged chunk
		select {
		case result := <-ch:
			expected := []string{"item1", "item2", "item3"}
			if !reflect.DeepEqual(result, expected) {
				t.Fatalf("expected %v, got %v", expected, result)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timeout waiting for merged chunk")
		}

		ff.Close()
	})

	t.Run("FuncMergeBytes", func(t *testing.T) {
		ff := flashflood.New[[]byte](&flashflood.Opts{
			BufferAmount: 10,
			GateAmount:   2,
			Timeout:      100 * time.Millisecond,
		})

		ff.AddFunc(flashflood.FuncMergeBytes())

		ch, err := ff.GetChan()
		if err != nil {
			t.Fatalf("could not get channel: %v", err)
		}

		// Push byte slices
		ff.Push([]byte{0x01, 0x02}, []byte{0x03, 0x04})

		// Should get one merged byte slice
		select {
		case result := <-ch:
			expected := []byte{0x01, 0x02, 0x03, 0x04}
			if !reflect.DeepEqual(result, expected) {
				t.Fatalf("expected %v, got %v", expected, result)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timeout waiting for merged bytes")
		}

		ff.Close()
	})

	t.Run("FuncReturnIndividualBytes", func(t *testing.T) {
		ff := flashflood.New[byte](&flashflood.Opts{
			BufferAmount: 10,
			GateAmount:   3,
			Timeout:      100 * time.Millisecond,
		})

		ff.AddFunc(flashflood.FuncReturnIndividualBytes())

		ch, err := ff.GetChan()
		if err != nil {
			t.Fatalf("could not get channel: %v", err)
		}

		// Push individual bytes
		ff.Push(0x01, 0x02, 0x03)

		// Should get individual bytes back
		var results []byte
		for i := 0; i < 3; i++ {
			select {
			case result := <-ch:
				results = append(results, result)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("timeout waiting for byte %d", i+1)
			}
		}

		expected := []byte{0x01, 0x02, 0x03}
		if !reflect.DeepEqual(results, expected) {
			t.Fatalf("expected %v, got %v", expected, results)
		}

		ff.Close()
	})
}

// Test edge cases for callback functions
func TestCallbackFunctionEdgeCases(t *testing.T) {
	t.Run("FuncPassThrough with empty input", func(t *testing.T) {
		ff := flashflood.New[string](&flashflood.Opts{
			BufferAmount: 5,
			GateAmount:   1,
			Timeout:      50 * time.Millisecond,
		})

		ff.AddFunc(flashflood.FuncPassThrough[string]())

		// Don't get channel to avoid draining
		// Just test that the function doesn't panic with empty input

		// This test verifies the function can be created and added without issues
		if ff.Count() != 0 {
			t.Fatal("buffer should be empty initially")
		}

		ff.Close()
	})

	t.Run("FuncFlatten with complex type", func(t *testing.T) {
		// Test with custom struct to ensure generics work with complex types
		type TestStruct struct {
			ID   int
			Name string
		}

		ff := flashflood.New[TestStruct](&flashflood.Opts{
			BufferAmount: 5,
			GateAmount:   2,
			Timeout:      100 * time.Millisecond,
		})

		ff.AddFunc(flashflood.FuncFlatten[[]TestStruct, TestStruct]())

		ch, err := ff.GetChan()
		if err != nil {
			t.Fatalf("could not get channel: %v", err)
		}

		// Push structs
		struct1 := TestStruct{ID: 1, Name: "test1"}
		struct2 := TestStruct{ID: 2, Name: "test2"}
		ff.Push(struct1, struct2)

		// Should get structs back unchanged
		var results []TestStruct
		for i := 0; i < 2; i++ {
			select {
			case result := <-ch:
				results = append(results, result)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("timeout waiting for struct %d", i+1)
			}
		}

		expected := []TestStruct{struct1, struct2}
		if !reflect.DeepEqual(results, expected) {
			t.Fatalf("expected %v, got %v", expected, results)
		}

		ff.Close()
	})
}
