package fs

import (
	"context"
	"io"
)

// A ProgressFunc is used to track progress of an operation.
type ProgressFunc = func(add int)

var progressFuncKey struct{}

// GetProgressFunc gets a progress function from a context. If no function is defined a no-op function is returned.
func GetProgressFunc(ctx context.Context) ProgressFunc {
	f, ok := ctx.Value(progressFuncKey).(ProgressFunc)
	if !ok {
		f = func(add int) {}
	}
	return f
}

// WithProgressFunc sets a progress function on a context.
func WithProgressFunc(ctx context.Context, f ProgressFunc) context.Context {
	return context.WithValue(ctx, progressFuncKey, f)
}

type progressReader struct {
	reader   io.Reader
	progress ProgressFunc
}

// NewProgressReader creates a new io.Reader which reports progress.
func NewProgressReader(r io.Reader, f ProgressFunc) io.Reader {
	return progressReader{
		reader:   r,
		progress: f,
	}
}

func (rdr progressReader) Read(p []byte) (n int, err error) {
	n, err = rdr.reader.Read(p)
	rdr.progress(n)
	return n, err
}
