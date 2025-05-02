package generics

func First[T any](items []T, cb func(item T) bool) *T {
	for _, item := range items {
		if cb(item) {
			return &item
		}
	}

	return nil
}
