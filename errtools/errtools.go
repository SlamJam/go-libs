package errtools

import "github.com/pkg/errors"

func NewOrWrap(err error, msg string) error {
	if err == nil {
		return errors.New(msg)
	}

	return errors.Wrap(err, msg)
}
