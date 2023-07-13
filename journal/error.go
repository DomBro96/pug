package journal

import (
	"errors"
	"fmt"
)

var _ error = (*ErrCorrupted)(nil)

// ErrCorrupted is the error type that generated by corrupted block or chunk.
type ErrCorrupted struct {
	Size   int
	Reason string
}

func (e *ErrCorrupted) Error() string {
	return fmt.Sprintf("pug/journal: block/chunk corrupted: %s (%d bytes)", e.Reason, e.Size)
}

var errSkip = errors.New("pug/journal: skipped")