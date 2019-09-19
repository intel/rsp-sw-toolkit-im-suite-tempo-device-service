package main

import "io"

// NewSkipSpaceReader returns a reader that skips newlines and spaces in an
// input stream.
func NewSpaceSkipReader(r io.Reader) *SpaceSkipReader {
	return &SpaceSkipReader{r: r}
}

// SkipSpaceReader wraps a reader, skipping newlines and spaces (0x20).
type SpaceSkipReader struct {
	r io.Reader
}

// Read reads bytes from the underlying reader and puts them in the buffer,
// skipping newlines and spaces along the way. It returns the number of bytes
// that weren't skipped along with the first error encountered, if any.
func (r *SpaceSkipReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	dst := 0
	for _, b := range p {
		if b == '\n' || b == ' ' {
			n--
			continue
		}
		p[dst] = b
		dst++
	}
	return
}
