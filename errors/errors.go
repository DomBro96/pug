package errors

import (
	"errors"
)

var _ error = (*ErrCorrupted)(nil)

// New returns an error that formats as the given text.
func New(text string) error {
	return errors.New(text)
}

// ErrCorrupted is the type that wraps errors that indicate corruption in
// the database.
type ErrCorrupted struct {
	//Fd  storage.FileDesc
	Err error
}

func (e *ErrCorrupted) Error() string {
	// if !e.Fd.Zero() {
	// 	return fmt.Sprintf("%v [file=%v]", e.Err, e.Fd)
	// }
	return e.Err.Error()
}
