package util

import "io"

func TryCloseReader(r io.Reader) error {
	if c, ok := r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func TryCloseWriter(w io.Writer) error {
	if c, ok := w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
