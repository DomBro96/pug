package journal

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/dombro/pug/errors"
	"github.com/dombro/pug/util"
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
	checksumStart = 0
	checksumEnd   = 4
	lengthStart   = 4
	lengthEnd     = 6
	chunkType     = 6
)

const (
	fullChunkType = iota + 1
	firstChunkType
	middleChunkType
	lastChunkType
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

func (r *Reader) corrupt(n int, reason string, skip bool) error {
	if r.dropper != nil {
		r.dropper.Drop(&ErrCorrupted{Size: n, Reason: reason})
	}

	if !skip && r.strict {
		r.err = &errors.ErrCorrupted{Err: &ErrCorrupted{Size: n, Reason: reason}}
		return r.err
	}

	return errSkip
}

// nextChunk sets r.buf[r.i:r.j] to hold the next chunk's payload, reading the
// next block into the buffer if necessary.
func (r *Reader) nextChunk(first bool) error {
	for {
		if r.j+headerSize <= r.n {
			checksum := binary.LittleEndian.Uint32(r.buf[r.j+checksumStart : r.j+checksumEnd])
			length := binary.LittleEndian.Uint16(r.buf[r.j+lengthStart : r.j+lengthEnd])
			chunkType := r.buf[r.j+chunkType]
			unprocBlock := r.n - r.j

			if checksum == 0 && length == 0 && chunkType == 0 {
				// Drop entire block.
				r.i = r.n
				r.j = r.n
				return r.corrupt(unprocBlock, "zero header", false)
			}

			if chunkType < fullChunkType || chunkType > lastChunkType {
				// Drop entire block.
				r.i = r.n
				r.j = r.n
				return r.corrupt(unprocBlock, fmt.Sprintf("invalid chunk type %#x", chunkType), false)
			}

			r.i = r.j + headerSize
			r.j = r.j + headerSize + int(length)

			if r.j > r.n {
				// Drop entire block.
				r.i = r.n
				r.j = r.n
				return r.corrupt(unprocBlock, "chunk length overflows block", false)
			}

			if r.checksum && checksum != util.NewCRC(r.buf[r.i-1:r.j]).Value() {
				// Drop entire block.
				r.i = r.n
				r.j = r.n
				return r.corrupt(unprocBlock, "checksum mismatch", false)
			}

			if first && chunkType != fullChunkType && chunkType != firstChunkType {
				chunkLength := (r.j - r.i) + headerSize
				r.i = r.j
				// Report the error, but skip it.
				return r.corrupt(chunkLength, "orphan chunk", true)
			}

			r.last = chunkType == fullChunkType || chunkType == lastChunkType
			return nil
		}

		// The last block.
		if r.n < blockSize && r.n > 0 {
			if !first {
				return r.corrupt(0, "missing chunk part", false)
			}
			r.err = io.EOF
			return r.err
		}

		// Read block.
		n, err := io.ReadFull(r.r, r.buf[:])
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		if n == 0 {
			if !first {
				return r.corrupt(0, "missing chunk part", false)
			}
			r.err = io.EOF
			return r.err
		}
		r.i, r.j, r.n = 0, 0, n
	}
}
