package journal

type flusher interface {
	Flush() error
}

// Dropper is the interface that wrap simple Drop method. The Drop
// method will be called when the journal reader dropping a block or chunk.
type Dropper interface {
	Drop(error)
}
