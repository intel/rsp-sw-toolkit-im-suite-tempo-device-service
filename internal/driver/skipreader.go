package driver

import "io"

// NewSpaceSkipReader returns a reader that skips newlines and spaces in an
// input stream.
func NewSpaceSkipReader(r io.Reader) *SpaceSkipReader {
	return &SpaceSkipReader{r: r}
}

// SpaceSkipReader wraps a reader, skipping line breaks, carriage returns, tabs,
// and spaces ('\n', '\r', '\t', and ' ').
type SpaceSkipReader struct {
	r io.Reader
}

// ignore these bytes in the byte stream
var ignore = [255]byte{'\n': 1, '\r': 1, '\t': 1, ' ': 1}

// Read reads bytes from the underlying reader and puts them in the buffer,
// skipping '\n', '\r, ' ', and '\t' along the way. It returns the number of
// bytes read but not skipped, along with the reader error encountered, if any.
//
// If all bytes are skipped, Read may return (0, nil); do NOT consider this EOF.
func (r *SpaceSkipReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	for i, dst, max := 0, 0, n; i < max; i++ {
		if ignore[p[i]] == 1 {
			n--
		} else {
			p[dst] = p[i]
			dst++
		}
	}
	return
}
