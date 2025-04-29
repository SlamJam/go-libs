package xgo

import (
	"fmt"
)

type PanicError struct {
	Payload any
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic was raised with payload: %+v", e.Payload)
}

var _ error = &PanicError{}

func PanicCatcher(f func()) (err *PanicError) {
	defer func() {
		if p := recover(); p != nil {
			err = &PanicError{Payload: p}
		}
	}()

	f()

	return nil
}

func PanicCatcherErr(f func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = &PanicError{Payload: p}
		}
	}()

	return f()
}

// func GoRecover(f func(), onPanic func(any)) {
// 	go func() {
// 		p := PanicCatcher(f)
// 		onPanic(p)
// 	}()
// }

func GoRecoverForever(f func()) <-chan any {
	out := make(chan any)

	go func() {
		defer close(out)
		for {
			p := PanicCatcher(f)
			if p == nil {
				break
			}

			out <- p
		}
	}()

	return out
}
