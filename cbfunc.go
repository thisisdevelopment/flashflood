package flashflood

// FuncMergeBytes returns function to merge all elements into one reply eg [1 2 3 4 5 6 7 8 9]
func (i *FlashFlood) FuncMergeBytes() FuncStack {

	f := func(objs []interface{}, ff *FlashFlood) []interface{} {
		var c []interface{}
		for _, v := range objs {
			iterable := v.([]uint8)
			for _, b := range iterable {
				c = append(c, b)
			}
		}
		var r []interface{}
		return append(r, c)
	}
	return f
}

//FuncReturnIndividualBytes returns function to merge all elements in containers into one reply eg  1 \n 2 ...
func (i *FlashFlood) FuncReturnIndividualBytes() FuncStack {
	f := func(objs []interface{}, ff *FlashFlood) []interface{} {
		var c []interface{}
		for _, v := range objs {
			iterable := v.([]uint8)
			for _, b := range iterable {
				c = append(c, b)
			}
		}
		return c
	}
	return f
}

//FuncMergeChunkedElements returns function to merge all elements in containers into one reply eg  [[1 2 3] [4 5 6] [7 8 9]]
func (i *FlashFlood) FuncMergeChunkedElements() FuncStack {
	f := func(objs []interface{}, ff *FlashFlood) []interface{} {
		var c []interface{}
		c = append(c, objs)
		return c
	}
	return f
}
