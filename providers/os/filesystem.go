// Package os implements a filesystem using the operating system.
package os

import (
	"context"
	"io"
	"net/url"
	"os"

	"github.com/badgerodon/fs"
)

func init() {
	fs.Register("file", New)
}

type fileSystem struct{}

// New creates a new FileSystem backed by the operating system.
func New() fs.FileSystem {
	return &fileSystem{}
}

func (fs *fileSystem) Create(ctx context.Context, dstURL *url.URL) (io.WriteCloser, error) {
	osf, err := os.Create(dstURL.Path)
	if err != nil {
		return nil, err
	}
	return &file{
		ctx:        ctx,
		underlying: osf,
	}, nil
}

func (fs *fileSystem) Open(ctx context.Context, srcURL *url.URL) (io.ReadCloser, error) {
	osf, err := os.Open(srcURL.Path)
	if err != nil {
		return nil, err
	}
	return &file{
		ctx:        ctx,
		underlying: osf,
	}, nil
}

func (fs *fileSystem) Stat(ctx context.Context, srcURL *url.URL) (fs.FileInfo, error) {
	fi, err := os.Stat(srcURL.Path)
	if err != nil {
		return nil, err
	}
	return fi, nil
}
