package std

// valise

// Пустое значение
type Void struct{}

func NewVoid() Void {
	return Void{}
}

func Zero[T any]() T {
	var t T
	return t
}

type Size interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64
}

func AssertSize[T Size](s T) {
	if s < 1 {
		panic("size must be gt 0")
	}
}
