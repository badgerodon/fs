package fs

import (
	"context"
	"io"
	"io/fs"
	"net/url"
)

// FileInfo is used for info about a file.
type FileInfo = fs.FileInfo

// A FileSystem implements file operations.
type FileSystem interface {
	Create(ctx context.Context, dstURL *url.URL) (io.WriteCloser, error)
	Open(ctx context.Context, srcURL *url.URL) (io.ReadCloser, error)
	Stat(ctx context.Context, srcURL *url.URL) (FileInfo, error)
}
