package flashflood

// FuncMergeChunkedElements returns function to merge individual element slices into a single chunk
// This is used with FlashFlood[[]T] where individual items are pushed as []T{item}
// Input: [][]T (batch of single-element slices), Output: [][]T (single chunk containing all items flattened)
func FuncMergeChunkedElements[T any]() FuncStack[[]T] {
	return func(objs [][]T, _ *FlashFlood[[]T]) [][]T {
		// Flatten all the single-element slices into one combined slice
		var flattened []T
		for _, slice := range objs {
			flattened = append(flattened, slice...)
		}
		// Return as a single chunk
		return [][]T{flattened}
	}
}

// FuncPassThrough returns function that passes elements through unchanged
// This is the default behavior - return items individually
func FuncPassThrough[T any]() FuncStack[T] {
	return func(objs []T, _ *FlashFlood[T]) []T {
		return objs
	}
}

// FuncFlatten flattens slice types into single elements
// Only works when T is a slice type like []byte, []string, etc.
func FuncFlatten[T ~[]E, E any]() FuncStack[E] {
	return func(objs []E, _ *FlashFlood[E]) []E {
		// This function works when the FlashFlood[E] receives flattened elements
		// The input objs are already individual E elements
		return objs
	}
}

// FuncMergeBytes merges multiple byte slices into a single combined byte slice
// Used with FlashFlood[[]byte] to merge individual byte slices
func FuncMergeBytes() FuncStack[[]byte] {
	return func(objs [][]byte, _ *FlashFlood[[]byte]) [][]byte {
		// Merge all byte slices into one
		var merged []byte
		for _, slice := range objs {
			merged = append(merged, slice...)
		}
		// Return as single merged slice
		return [][]byte{merged}
	}
}

// FuncReturnIndividualBytes flattens byte slices into individual byte values
// Used with FlashFlood[byte] to output individual bytes from byte slices
func FuncReturnIndividualBytes() FuncStack[byte] {
	return func(objs []byte, _ *FlashFlood[byte]) []byte {
		// Input is already individual bytes, return as-is
		return objs
	}
}
