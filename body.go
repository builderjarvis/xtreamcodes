package xtreamcodes

import (
	"context"
	"io"
)

// boundedBody measures decompressed bytes and fails instead of synthesizing EOF
// at the limit, which could make an oversized response look valid to a decoder.
type boundedBody struct {
	io.ReadCloser
	remaining int64
	ctx       context.Context
	err       error
}

func (b *boundedBody) Read(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if b.remaining == 0 {
		var extra [1]byte
		n, err := b.ReadCloser.Read(extra[:])
		if n > 0 {
			b.err = ErrResponseTooLarge
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > b.remaining {
		p = p[:b.remaining]
	}
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)

	return n, err
}
