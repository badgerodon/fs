package http

import (
	"context"
	"io"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
)

// A Download downloads data from an endpoint.
type Download struct {
	eg        *errgroup.Group
	ctx       context.Context
	pr        *io.PipeReader
	pw        *io.PipeWriter
	closeOnce sync.Once
	closeErr  error
}

// NewDownload creates a new download using the given http client and request.
func NewDownload(ctx context.Context, client *http.Client, req *http.Request) *Download {
	download := new(Download)
	download.eg, download.ctx = errgroup.WithContext(ctx)
	download.pr, download.pw = io.Pipe()
	download.eg.Go(func() error {
		req = req.WithContext(ctx)

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()

		if res.StatusCode/100 != 2 {
			return err
		}

		_, err = io.Copy(download.pw, res.Body)
		if err != nil {
			return err
		}

		return nil
	})
	return download
}

// Close closes the download.
func (download *Download) Close() error {
	download.closeOnce.Do(func() {
		if err := download.pw.Close(); err != nil {
			download.closeErr = err
		}
		if err := download.eg.Wait(); err != nil {
			download.closeErr = err
		}
	})
	return download.closeErr
}

// Read reads bytes from the HTTP response.
func (download *Download) Read(p []byte) (int, error) {
	return download.pr.Read(p)
}
