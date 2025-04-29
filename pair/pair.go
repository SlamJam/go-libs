package pair

type Pair[T1, T2 any] struct {
	First  T1
	Second T2
}

func New[T1, T2 any](t1 T1, t2 T2) Pair[T1, T2] {
	return Pair[T1, T2]{
		First:  t1,
		Second: t2,
	}
}
