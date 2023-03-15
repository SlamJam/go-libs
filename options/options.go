package options

type Opt[T any] func(T) T

func ApplyOptsInto[T any](in *T, opts ...Opt[T]) {
	o := *in
	for _, opt := range opts {
		o = opt(o)
	}

	*in = o
}
