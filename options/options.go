package options

type Opt[T any] func(*T)

func ApplyInto[T any](options *T, opts ...Opt[T]) {
	if options == nil {
		return
	}

	for _, opt := range opts {
		tmp := *options
		opt(&tmp)
		*options = tmp
	}
}

func Create[T any](opts ...Opt[T]) T {
	var options T
	ApplyInto(&options, opts...)
	return options
}
