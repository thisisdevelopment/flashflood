package flashflood_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/thisisdevelopment/flashflood"
)

func getTestObjs(amount int) []TestObj {
	r := []TestObj{}
	for i := 1; i <= amount; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		t := TestObj{key, value}
		r = append(r, t)
	}
	return r
}

//example using  Push method buffer of 3 pushing 5 elements, no gate
func ExampleFlashFlood_Push_example01() {

	ff := flashflood.New(&flashflood.Opts{
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
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	// push element in buffer
	ff.Push(o[4])
	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}
	// Output: KV: k1 v1
	// KV: k2 v2
}
func ExampleFlashFlood_Unshift_example01() {

	ff := flashflood.New(&flashflood.Opts{
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
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	// Output: KV: k4 v4
	// KV: k5 v5
	// KV: k1 v1
}

func ExampleFlashFlood_Unshift_example02() {

	ff := flashflood.New(&flashflood.Opts{
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
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	// Output: KV: k4 v4
	// KV: k5 v5
	// KV: k1 v1
}

func ExampleFlashFlood_Purge() {

	// instance
	ff := flashflood.New(&flashflood.Opts{
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
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	select {
	case v := <-ch:
		fmt.Println("KV:", v.(TestObj).Key, v.(TestObj).Value)
	}

	// Output: KV: k3 v3
	// KV: k4 v4
	// KV: k5 v5
}

//example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, use merge callback function to merge output. Put in artificial sleep to make sure timeout is reached
func ExampleFlashFlood_Push_example02() {

	ff := flashflood.New(&flashflood.Opts{
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

	// add function to callstack. FuncMergeBytes is a function that workes on byteslices, but you can use your own see example09
	ff.AddFunc(ff.FuncMergeBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	time.Sleep(500 * time.Millisecond)

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
	//RESULT: [10 11 12 13 14 15 16]
}

//example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 2, use merge callback function to merge output, not waiting for a timeout
func ExampleFlashFlood_Push_example03() {

	ff := flashflood.New(&flashflood.Opts{
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

	ff.AddFunc(ff.FuncMergeBytes())

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

//example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 2, use merge callback function to merge output, wait for timeout
func ExampleFlashFlood_Push_example04() {

	ff := flashflood.New(&flashflood.Opts{
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

	ff.AddFunc(ff.FuncMergeBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])
	time.Sleep(500 * time.Millisecond)

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
	//RESULT: [7 8 9 10 11 12]
	//RESULT: [13 14 15 16]
}

//example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, use merge callback function to merge output, not waiting for timeout (result is no output at all)
func ExampleFlashFlood_Push_example05() {

	ff := flashflood.New(&flashflood.Opts{
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

	ff.AddFunc(ff.FuncMergeBytes())

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

//example using Push method buffer of 4 pushing 6 byte slices of 3 bytes each, gate of 3, use callback function send individual bytes on channel, wait for timeout
func ExampleFlashFlood_Push_example06() {

	ff := flashflood.New(&flashflood.Opts{
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

	ff.AddFunc(ff.FuncReturnIndividualBytes())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5])

	time.Sleep(500 * time.Millisecond)

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
	//RESULT: 2
	//RESULT: 3
	//RESULT: 4
	//RESULT: 5
	//RESULT: 6
	//RESULT: 7
	//RESULT: 8
	//RESULT: 9
	//RESULT: 10
	//RESULT: 11
	//RESULT: 12
	//RESULT: 13
	//RESULT: 14
	//RESULT: 15
	//RESULT: 16
}

//example using Push method buffer of 3 pushing 7 byte slices of 3 bytes each, gate of 3, use callback function send elements on channel, wait for timeout
func ExampleFlashFlood_Push_example07() {

	ff := flashflood.New(&flashflood.Opts{
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

	ff.AddFunc(ff.FuncMergeChunkedElements())

	ch, err := ff.GetChan()
	_ = err
	ff.Push(b[0], b[1], b[2], b[3], b[4], b[5], b[6])

	time.Sleep(500 * time.Millisecond)

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

	//Output: RESULT: [[1 2 3] [4 5 6] [7 8 9]]
	//RESULT: [[10 11 12] [13 14 15] [16 17 18]]
	//RESULT: [[19]]
}

//example Debug enabled
func ExampleFlashFlood_Push_example08() {

	ff := flashflood.New(&flashflood.Opts{
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
			fmt.Println(v.(TestObj).Key)
		}
	}

	//Output: DEBUG: []interface {}{flashflood_test.TestObj{Key:"k1", Value:"v1"}, flashflood_test.TestObj{Key:"k2", Value:"v2"}}
	//k1
	//k2
	//DEBUG: []interface {}{flashflood_test.TestObj{Key:"k3", Value:"v3"}, flashflood_test.TestObj{Key:"k4", Value:"v4"}, flashflood_test.TestObj{Key:"k5", Value:"v5"}}
	//k3
	//k4
	//k5
}

//example Debug enabled using callback function.
func ExampleFlashFlood_Push_example09() {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
		Debug:        true,
	})

	o := getTestObjs(5)

	ch, _ := ff.GetChan()

	f := func(objs []interface{}, ff *flashflood.FlashFlood) []interface{} {
		var c []interface{}
		for k, v := range objs {
			c = append(c, fmt.Sprintf("---%d_%s", k, v))
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

	//Output: DEBUG: []interface {}{"---0_{k1 v1}", "---1_{k2 v2}"}
	//---0_{k1 v1}
	//---1_{k2 v2}
	//DEBUG: []interface {}{"---0_{k3 v3}", "---1_{k4 v4}", "---2_{k5 v5}"}
	//---0_{k3 v3}
	//---1_{k4 v4}
	//---2_{k5 v5}
}

//example Debug enabled using multiple callback function. Each function is performed on  the element flushed out to the channel in order
func ExampleFlashFlood_Push_example10() {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
		Debug:        true,
	})

	o := getTestObjs(5)

	ch, _ := ff.GetChan()

	f1 := func(objs []interface{}, ff *flashflood.FlashFlood) []interface{} {
		var c []interface{}
		for _, v := range objs {
			c = append(c, fmt.Sprintf("---%s_%s", v.(TestObj).Key, v.(TestObj).Value))
		}
		return c
	}

	f2 := func(objs []interface{}, ff *flashflood.FlashFlood) []interface{} {
		var c []interface{}
		for _, v := range objs {
			c = append(c, strings.Replace(v.(string), "v", "", -1))
		}
		return c
	}

	ff.AddFunc(f1)
	ff.AddFunc(f2)

	ff.Push(o[0], o[1], o[2], o[3], o[4])

	for i := 0; i < 5; i++ {
		select {
		case v := <-ch:
			fmt.Printf("%v\n", v)
		}
	}

	//Output: DEBUG: []interface {}{"---k1_1", "---k2_2"}
	//---k1_1
	//---k2_2
	//DEBUG: []interface {}{"---k3_3", "---k4_4", "---k5_5"}
	//---k3_3
	//---k4_4
	//---k5_5
}

//example to drain elements on channel
func ExampleFlashFlood_Drain_example01() {

	ff := flashflood.New(&flashflood.Opts{
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

	//Output: RESULT: [1 2 3]
	//RESULT: [4 5 6]
	//RESULT: [7 8 9]
	//RESULT: [10 11 12]
	//RESULT: [13 14 15]
	//RESULT: [16 17 18]
}

//example to drain elements not using the channel and return them
func ExampleFlashFlood_Drain_example02() {

	ff := flashflood.New(&flashflood.Opts{
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

	//Output: RESULT: [[10 11 12] [13 14 15] [16 17 18]]
}
