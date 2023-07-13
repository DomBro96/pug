package journal

import (
	"io"

	"github.com/dombro/pug/errors"
)

// The wire format is that the stream is divided into 32KiB blocks, and each
// block contains a number of tightly packed chunks. Chunks cannot cross block
// boundaries. The last block may be shorter than 32 KiB. Any unused bytes in a
// block must be zero.
// A journal maps to one or more chunks. Each chunk has a 7 byte header (a 4
// byte checksum, a 2 byte little-endian uint16 length, and a 1 byte chunk type)
// followed by a payload. The checksum is over the chunk type and the payload.
//
// There are four chunk types: whether the chunk is the full journal, or the
// first, middle or last chunk of a multi-chunk journal. A multi-chunk journal
// has one first chunk, zero or more middle chunks, and one last chunk.

const (
	blockSize  = 32 * 1024
	headerSize = 7
)

const (
	fullChunckType = iota + 1
	firstChunckType
	middleChunckType
	lastChunckType
)

//  Reader reads journals from an underlying io.Reader.
type Reader struct {
	// r is the underlying reader.
	r io.Reader

	dropper Dropper
	// strict flag.
	strict bool
	// checksum flag.
	checksum bool
	// seq is the sequence number of the current journal.
	seq int
	// buf[i:j] is the unread portion of the current chunk's payload.
	// The low bound, i, excludes the chunk header.
	i, j int
	// n is the number of bytes of buf that are valid. Once reading has started,
	// only the final block can have n < blockSize.
	n int
	// last is whether the current chunk is the last chunk of the journal.
	last bool
	// err is any accumulated error.
	err error
	// buf is the buffer.
	buf [blockSize]byte
}

func NewReader(r io.Reader, dropper Dropper, strict bool, checksum bool) *Reader {
	return &Reader{
		r:        r,
		dropper:  dropper,
		strict:   strict,
		checksum: checksum,
		last:     true,
	}
}

func (r *Reader) corrupted(n int, reason string, skip bool) error {
	if r.dropper != nil {
		r.dropper.Drop(&ErrCorrupted{Size: n, Reason: reason})
	}

	if !skip && r.strict {
		r.err = &errors.ErrCorrupted{Err: &ErrCorrupted{Size: n, Reason: reason}}
		return r.err
	}

	return errSkip
}
