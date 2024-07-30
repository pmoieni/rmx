package lib

import "errors"

func Pipe[T any](in T, list ...func(T) error) error {
	var errs []error

	for _, f := range list {
		errs = append(errs, f(in))
	}

	return errors.Join(errs...)
}
