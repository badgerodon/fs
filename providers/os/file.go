package os

import (
	"context"
	"os"
)

type file struct {
	ctx        context.Context
	underlying *os.File
}

func (f *file) Close() error {
	return f.underlying.Close()
}

func (f *file) Read(p []byte) (int, error) {
	select {
	case <-f.ctx.Done():
		return 0, f.ctx.Err()
	default:
	}
	return f.underlying.Read(p)
}

func (f *file) Write(p []byte) (int, error) {
	select {
	case <-f.ctx.Done():
		return 0, f.ctx.Err()
	default:
	}
	return f.underlying.Write(p)
}
