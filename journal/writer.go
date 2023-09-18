package journal

import (
	"encoding/binary"
	"io"

	"github.com/dombro/pug/util"
)

// Writer writes journals to an underlying io.Writer.
type Writer struct {
	// w is the underlying writer.
	w io.Writer
	// seq is the sequence number of the current journal.
	seq int
	// f is w as a flusher.
	f flusher
	// buf[i:j] is the bytes that will become the current chunk.
	// The low bound, i, includes the chunk header.
	i, j int
	// buf[:written] has already been written to w.
	// written is zero unless Flush has been called.
	written int
	// first is whether the current chunk is the first chunk of the journal.
	first bool
	// pending is whether a chunk is buffered but not yet written.
	pending bool
	// err is any accumulated error.
	err error
	// buf is the buffer.
	buf [blockSize]byte
}

func NewWriter(w io.Writer) *Writer {
	f, _ := w.(flusher)
	return &Writer{
		w: w,
		f: f,
	}
}

// fillHeader fills in the header for the pending chunk.
func (w *Writer) fillHeader(last bool) {
	if w.i+headerSize > w.j || w.j > blockSize {
		panic("pug/journal: bad writer state")
	}
	if last {
		if w.first {
			w.buf[w.i+idxChunkType] = chunkTypeFull
		} else {
			w.buf[w.i+idxChunkType] = chunkTypeLast
		}
	} else {
		if w.first {
			w.buf[w.i+6] = chunkTypeFirst
		} else {
			w.buf[w.i+6] = chunkTypeMiddle
		}
	}

	binary.LittleEndian.PutUint32(w.buf[w.i+idxChecksumStart:w.i+idxChecksumEnd], util.NewCRC(w.buf[w.i+idxLengthStart:w.j]).Value())
	binary.LittleEndian.PutUint16(w.buf[w.i+idxLengthStart:w.i+idxLengthEnd], uint16(w.j-w.i-headerSize))
}
