package comparer

// BasicComparer is the interface that wraps the basic Compare method.
type BasicComparer interface {
	// Compare return -1, 0, or +1 depending on whether a is 'less than',
	// 'equal to' or 'garter than' b. The two args can only be 'equal' if
	// their contents are exactly equal. Furthermore the empty slice must
	// be 'less than' any none-empty slice.
	Compare(a, b []byte) int
}

type Comparer interface {
	BasicComparer

	// Name returns name of the comparer.
	Name() string

	// Bellow are advanced functions.

	Separator(dst, a, b []byte) []byte

	Successor(dst, b []byte) []byte
}
