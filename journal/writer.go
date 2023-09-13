package journal

import "io"

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
