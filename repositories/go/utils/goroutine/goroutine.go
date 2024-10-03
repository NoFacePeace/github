package goroutine

func RecoverGo(f func()) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
}
