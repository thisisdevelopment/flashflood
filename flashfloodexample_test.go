package flashflood_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/thisisdevelopment/flashflood/v2"
)

// TestObj and getTestObjs are defined in flashflood__01_test.go

// example using  Push method buffer of 3 pushing 5 elements, no gate
func ExampleFlashFlood_Push_example01() {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})

	// o[0] := TestObj{Key: "k1", Value: "v1"}
	// ...
	o := getTestObjs(5)

	// get the channel from buffer
	ch, err := ff.GetChan()
	_ = err
	// push elements in buffer
	ff.Push(o[0], o[1], o[2], o[3])

	// listen on channel
	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	// push element in buffer
	ff.Push(o[4])
	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}
	// Output: KV: k1 v1
	// KV: k2 v2
}

func ExampleFlashFlood_Unshift_example01() {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})

	o := getTestObjs(5)

	ch, err := ff.GetChan()
	_ = err
	ff.Push(o[0], o[1], o[2])

	ff.Unshift(o[3])
	ff.Unshift(o[4])

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	// Output: KV: k4 v4
	// KV: k5 v5
	// KV: k1 v1
}

func ExampleFlashFlood_Unshift_example02() {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})

	o := getTestObjs(5)

	ch, err := ff.GetChan()
	_ = err
	ff.Push(o[0], o[1], o[2])
	ff.Unshift(o[3], o[4])

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	// Output: KV: k4 v4
	// KV: k5 v5
	// KV: k1 v1
}

func ExampleFlashFlood_Purge() {
	// instance
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 2,
		Timeout:      time.Duration(250 * time.Millisecond),
	})

	o := getTestObjs(5)

	// Prep test Objs

	// get the channel
	ch, err := ff.GetChan()
	_ = err

	// push to buffer
	ff.Push(o[0], o[1])

	// purge buffer
	ff.Purge()

	// push to buffer
	ff.Push(o[2], o[3], o[4])

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.Key, v.Value)
	}

	// Output: KV: k3 v3
	// KV: k4 v4
	// KV: k5 v5
}

// example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, deprecated byte functions removed in v2
func ExampleFlashFlood_Push_example02() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 4,
		Timeout:      time.Duration(100 * time.Millisecond), // Reduced from 250ms for CI
		GateAmount:   3,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10},
	}

	// Add byte merging function
	ff.AddFunc(flashflood.FuncMergeBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	time.Sleep(200 * time.Millisecond) // Reduced from 500ms for CI

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: [1 2 3 4 5 6 7 8 9]
	// RESULT: [10 11 12 13 14 15 16]
}

// example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 2, use merge callback function to merge output, not waiting for a timeout
func ExampleFlashFlood_Push_example03() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 4,
		Timeout:      time.Duration(250 * time.Millisecond),
		GateAmount:   2,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10},
	}

	// Add byte merging function
	ff.AddFunc(flashflood.FuncMergeBytes())

	ch, err := ff.GetChan()

	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: [1 2 3 4 5 6]
}

// example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 2, use merge callback function to merge output, wait for timeout
func ExampleFlashFlood_Push_example04() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 4,
		Timeout:      time.Duration(100 * time.Millisecond), // Reduced from 250ms for CI
		GateAmount:   2,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10},
	}

	// Add byte merging function
	ff.AddFunc(flashflood.FuncMergeBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])
	time.Sleep(200 * time.Millisecond) // Reduced from 500ms for CI

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: [1 2 3 4 5 6]
	// RESULT: [7 8 9 10 11 12]
	// RESULT: [13 14 15 16]
}

// example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, use merge callback function to merge output, not waiting for timeout (result is no output at all)
func ExampleFlashFlood_Push_example05() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 4,
		Timeout:      time.Duration(250 * time.Millisecond),
		GateAmount:   3,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10},
	}

	// Add byte merging function
	ff.AddFunc(flashflood.FuncMergeBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output:
}

// example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, use callback function send individual bytes on channel, wait for timeout
func ExampleFlashFlood_Push_example06() {
	ff := flashflood.New[byte](&flashflood.Opts{
		BufferAmount: 4,
		Timeout:      time.Duration(100 * time.Millisecond), // Reduced from 250ms for CI
		GateAmount:   3,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10},
	}

	// Add individual bytes function
	ff.AddFunc(flashflood.FuncReturnIndividualBytes())

	ch, err := ff.GetChan()
	_ = err

	// Push individual bytes from the byte slices
	for _, slice := range b {
		for _, byteVal := range slice {
			ff.Push(byteVal)
		}
	}

	time.Sleep(200 * time.Millisecond) // Reduced from 500ms for CI

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: 1
	// RESULT: 2
	// RESULT: 3
	// RESULT: 4
	// RESULT: 5
	// RESULT: 6
	// RESULT: 7
	// RESULT: 8
	// RESULT: 9
	// RESULT: 10
	// RESULT: 11
	// RESULT: 12
	// RESULT: 13
	// RESULT: 14
	// RESULT: 15
	// RESULT: 16
}

// example using Push method buffer of 3 pushing 7 byte slices of 3 bytes each, gate of 3, use callback function send elements on channel, wait for timeout
func ExampleFlashFlood_Push_example07() {
	ff := flashflood.New[[][]byte](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(100 * time.Millisecond),
		GateAmount:   3,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10, 0x11, 0x12},
		{0x13},
	}

	ff.AddFunc(flashflood.FuncMergeChunkedElements[[]byte]())

	ch, err := ff.GetChan()
	_ = err
	// Push individual byte slices for chunked processing
	for _, bs := range b {
		ff.Push([][]byte{bs}) // Wrap each []byte in [][]byte for chunking
	}

	time.Sleep(200 * time.Millisecond) // Reduced from 500ms for CI

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: [[1 2 3] [4 5 6] [7 8 9]]
	// RESULT: [[10 11 12] [13 14 15] [16 17 18]]
	// RESULT: [[19]]
}

// example Debug enabled
func ExampleFlashFlood_Push_example08() {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
		Debug:        true,
	})

	o := getTestObjs(5)

	ch, _ := ff.GetChan()

	ff.Push(o[0], o[1], o[2], o[3], o[4])

	for i := 0; i < 5; i++ {
		select {
		case v := <-ch:
			fmt.Println(v.Key)
		}
	}

	// Output: DEBUG: []flashflood_test.TestObj{flashflood_test.TestObj{Key:"k1", Value:"v1"}, flashflood_test.TestObj{Key:"k2", Value:"v2"}}
	// k1
	// k2
	// DEBUG: []flashflood_test.TestObj{flashflood_test.TestObj{Key:"k3", Value:"v3"}, flashflood_test.TestObj{Key:"k4", Value:"v4"}, flashflood_test.TestObj{Key:"k5", Value:"v5"}}
	// k3
	// k4
	// k5
}

// example Debug enabled using callback function.
func ExampleFlashFlood_Push_example09() {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
		Debug:        true,
	})

	o := getTestObjs(5)

	ch, _ := ff.GetChan()

	f := func(objs []TestObj, _ *flashflood.FlashFlood[TestObj]) []TestObj {
		var c []TestObj
		for k, v := range objs {
			// Transform TestObj to new TestObj with modified fields
			transformed := TestObj{
				Key:   fmt.Sprintf("---%d_%s", k, v.Key),
				Value: v.Value,
			}
			c = append(c, transformed)
		}
		return c
	}

	ff.AddFunc(f)

	ff.Push(o[0], o[1], o[2], o[3], o[4])

	for i := 0; i < 5; i++ {
		select {
		case v := <-ch:
			fmt.Println(v)
		}
	}

	//Output: DEBUG: []flashflood_test.TestObj{flashflood_test.TestObj{Key:"---0_k1", Value:"v1"}, flashflood_test.TestObj{Key:"---1_k2", Value:"v2"}}
	//{---0_k1 v1}
	//{---1_k2 v2}
	//DEBUG: []flashflood_test.TestObj{flashflood_test.TestObj{Key:"---0_k3", Value:"v3"}, flashflood_test.TestObj{Key:"---1_k4", Value:"v4"}, flashflood_test.TestObj{Key:"---2_k5", Value:"v5"}}
	//{---0_k3 v3}
	//{---1_k4 v4}
	//{---2_k5 v5}
}

// example Debug enabled using multiple callback function. Each function is performed on  the element flushed out to the channel in order
func ExampleFlashFlood_Push_example10() {
	ff := flashflood.New[string](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
		Debug:        true,
	})

	o := getTestObjs(5)

	ch, _ := ff.GetChan()

	// Convert TestObj to string and apply transformations
	f1 := func(objs []string, _ *flashflood.FlashFlood[string]) []string {
		var c []string
		for _, v := range objs {
			// First transformation: add prefix
			c = append(c, fmt.Sprintf("---%s", v))
		}
		return c
	}

	f2 := func(objs []string, _ *flashflood.FlashFlood[string]) []string {
		var c []string
		for _, v := range objs {
			// Second transformation: remove 'v'
			c = append(c, strings.ReplaceAll(v, "v", ""))
		}
		return c
	}

	ff.AddFunc(f1)
	ff.AddFunc(f2)

	// Convert TestObj to strings and push
	ff.Push(
		fmt.Sprintf("%s_%s", o[0].Key, o[0].Value),
		fmt.Sprintf("%s_%s", o[1].Key, o[1].Value),
		fmt.Sprintf("%s_%s", o[2].Key, o[2].Value),
		fmt.Sprintf("%s_%s", o[3].Key, o[3].Value),
		fmt.Sprintf("%s_%s", o[4].Key, o[4].Value),
	)

	for i := 0; i < 5; i++ {
		select {
		case v := <-ch:
			fmt.Printf("%v\n", v)
		}
	}

	//Output: DEBUG: []string{"---k1_1", "---k2_2"}
	//---k1_1
	//---k2_2
	//DEBUG: []string{"---k3_3", "---k4_4", "---k5_5"}
	//---k3_3
	//---k4_4
	//---k5_5
}

// example to drain elements on channel
func ExampleFlashFlood_Drain_example01() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(100 * time.Millisecond),
		GateAmount:   2,
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10, 0x11, 0x12},
	}

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	ff.Drain(true, false)

	run := true
	for run {
		select {
		case v := <-ch:
			fmt.Println("RESULT:", v)
		default:
			run = false
			break
		}
	}

	// Output: RESULT: [1 2 3]
	// RESULT: [4 5 6]
	// RESULT: [7 8 9]
	// RESULT: [10 11 12]
	// RESULT: [13 14 15]
	// RESULT: [16 17 18]
}

// example to drain elements not using the channel and return them
func ExampleFlashFlood_Drain_example02() {
	ff := flashflood.New[[]byte](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(100 * time.Millisecond),
	})

	b := [][]byte{
		{0x1, 0x2, 0x3},
		{0x4, 0x5, 0x6},
		{0x7, 0x8, 0x9},
		{0xa, 0xb, 0xc},
		{0xd, 0xe, 0xf},
		{0x10, 0x11, 0x12},
	}

	ch, err := ff.GetChan()
	_ = err
	_ = ch
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	e, _ := ff.Drain(false, false)

	fmt.Println("RESULT:", e)

	// Output: RESULT: [[10 11 12] [13 14 15] [16 17 18]]
}
