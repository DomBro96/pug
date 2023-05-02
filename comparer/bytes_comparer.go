package comparer

import "bytes"

var _ Comparer = (*BytesComparer)(nil)

type BytesComparer struct {
}

func (c *BytesComparer) Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func (c *BytesComparer) Name() string {
	return "pug.bytescomparator"
}

func (c *BytesComparer) Separator(dst, a, b []byte) []byte {
	return nil
}

func (c *BytesComparer) Successor(dst, b []byte) []byte {
	return nil
}