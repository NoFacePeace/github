package mini

import "testing"

func TestNewSlice(t *testing.T) {
	stackCandidate := NewSlice(3, 4)
	if stackCandidate.array == nil {
		t.Fatal("NewSlice() stack candidate has nil array")
	}
	if stackCandidate.len != 3 || stackCandidate.cap != 4 {
		t.Fatalf("NewSlice() = len %d, cap %d; want len 3, cap 4", stackCandidate.len, stackCandidate.cap)
	}

	heapPath := NewSlice(3, 5)
	if heapPath.array == nil {
		t.Fatal("NewSlice() heap path has nil array")
	}
	if heapPath.len != 3 || heapPath.cap != 5 {
		t.Fatalf("NewSlice() = len %d, cap %d; want len 3, cap 5", heapPath.len, heapPath.cap)
	}
}

func TestNewSlicePanicsWhenLengthExceedsCapacity(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewSlice() did not panic")
		}
	}()

	NewSlice(6, 5)
}

func TestAppendSlice(t *testing.T) {
	source := NewSlice(3, 4)
	withoutGrowth := AppendSlice(source, []int{4})

	if withoutGrowth.array != source.array {
		t.Fatal("AppendSlice() allocated when capacity was sufficient")
	}
	if withoutGrowth.len != 4 || withoutGrowth.cap != 4 {
		t.Fatalf("AppendSlice() = len %d, cap %d; want len 4, cap 4", withoutGrowth.len, withoutGrowth.cap)
	}

	withGrowth := AppendSlice(withoutGrowth, []int{5})
	if withGrowth.array == withoutGrowth.array {
		t.Fatal("AppendSlice() did not allocate when capacity was insufficient")
	}
	if withGrowth.len != 5 || withGrowth.cap != 8 {
		t.Fatalf("AppendSlice() = len %d, cap %d; want len 5, cap 8", withGrowth.len, withGrowth.cap)
	}
}

func TestAppendSliceLargeGrowth(t *testing.T) {
	source := NewSlice(256, 256)
	got := AppendSlice(source, []int{256})

	if got.len != 257 || got.cap != 512 {
		t.Fatalf("AppendSlice() = len %d, cap %d; want len 257, cap 512", got.len, got.cap)
	}
}
