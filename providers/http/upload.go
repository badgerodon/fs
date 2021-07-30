package http

import (
	"context"
	"io"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
)

// An Upload uploads data to an HTTP endpoint.
type Upload struct {
	eg        *errgroup.Group
	ctx       context.Context
	pr        *io.PipeReader
	pw        *io.PipeWriter
	closeOnce sync.Once
	closeErr  error
}

// NewUpload creates a new upload using the given http client and request.
//
// The request body will be overwritten with a pipe reader used by `Write`.
// `Close` will complete the upload.
func NewUpload(ctx context.Context, client *http.Client, req *http.Request) *Upload {
	upload := new(Upload)
	upload.eg, upload.ctx = errgroup.WithContext(ctx)
	upload.pr, upload.pw = io.Pipe()
	upload.eg.Go(func() error {
		req = req.WithContext(ctx)
		req.Body = upload.pr

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()

		if res.StatusCode/100 != 2 {
			return err
		}

		_, _ = io.ReadAll(res.Body)

		return nil
	})
	return upload
}

func (upload *Upload) Close() error {
	upload.closeOnce.Do(func() {
		if err := upload.pw.Close(); err != nil {
			upload.closeErr = err
		}
		if err := upload.eg.Wait(); err != nil {
			upload.closeErr = err
		}
	})
	return upload.closeErr
}

func (upload *Upload) Write(p []byte) (int, error) {
	return upload.pw.Write(p)
}
