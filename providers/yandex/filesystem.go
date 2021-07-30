package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"time"

	"github.com/badgerodon/fs"
	"github.com/badgerodon/fs/providers/http"
)

func init() {
	fs.Register("yandex", func() fs.FileSystem { return New() })
}

type fileSystem struct {
	cfg *config
}

// New creates a new FileSystem using Yandex.Disk.
func New(opts ...Option) fs.FileSystem {
	f := &fileSystem{cfg: getConfig(opts...)}
	return f
}

func (f *fileSystem) Create(ctx context.Context, dstURL *url.URL) (io.WriteCloser, error) {
	method, resourceURL, err := f.getResourceLink(ctx, "upload", dstURL)
	if err != nil {
		return nil, err
	}

	req, err := f.getRequest(ctx, method, resourceURL)
	if err != nil {
		return nil, err
	}

	return http.NewUpload(ctx, f.cfg.httpClient, req), nil
}

func (f *fileSystem) Open(ctx context.Context, srcURL *url.URL) (io.ReadCloser, error) {
	method, resourceURL, err := f.getResourceLink(ctx, "download", srcURL)
	if err != nil {
		return nil, err
	}

	req, err := f.getRequest(ctx, method, resourceURL)
	if err != nil {
		return nil, err
	}

	return http.NewDownload(ctx, f.cfg.httpClient, req), nil
}

func (f *fileSystem) Stat(ctx context.Context, srcURL *url.URL) (fs.FileInfo, error) {
	qs := srcURL.Query()
	qs.Set("path", srcURL.Path)

	baseURL, err := url.Parse(f.cfg.baseURL)
	if err != nil {
		return nil, err
	}
	srcURL = baseURL.ResolveReference(&url.URL{
		Path:     "/v1/disk/resources",
		RawQuery: qs.Encode(),
	})
	req, err := f.getRequest(ctx, "GET", srcURL)
	if err != nil {
		return nil, err
	}

	res, err := f.cfg.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return nil, err
	}

	var metadata struct {
		Name     string    `json:"name"`
		Size     int64     `json:"size"`
		Modified time.Time `json:"modified"`
	}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		name:    metadata.Name,
		size:    metadata.Size,
		modTime: metadata.Modified,
	}, nil
}

func (f *fileSystem) getRequest(ctx context.Context, method string, u *url.URL) (*stdhttp.Request, error) {
	qs := u.Query()

	accessToken := qs.Get("access_token")
	if accessToken != "" {
		qs.Del("access_token")
		u.RawQuery = qs.Encode()
	} else {
		accessToken = f.cfg.accessToken
	}

	req, err := stdhttp.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)
	return req, err
}

func (f *fileSystem) getResourceLink(ctx context.Context, action string, u *url.URL) (method string, resourceURL *url.URL, err error) {
	// u == yandex:///some/path?access_token=xyz123
	qs := u.Query()
	qs.Set("path", u.Path)
	qs.Set("overwrite", "true")

	baseURL, err := url.Parse(f.cfg.baseURL)
	if err != nil {
		return "", nil, err
	}
	u = baseURL.ResolveReference(&url.URL{
		Path:     "/v1/disk/resources/" + action,
		RawQuery: qs.Encode(),
	})

	req, err := f.getRequest(ctx, "GET", u)
	if err != nil {
		return "", nil, fmt.Errorf("yandex: error building %s request: %w", action, err)
	}

	res, err := f.cfg.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("yandex: error retrieving %s URL: %w", action, err)
	}
	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		return "", nil, fmt.Errorf("yandex: error reading %s response body: %w", action, err)
	}

	if res.StatusCode/100 != 2 {
		return "", nil, fmt.Errorf("yandex: unexpected %s status code: %d", action, res.StatusCode)
	}

	var link struct {
		HREF   string `json:"href"`
		Method string `json:"method"`
	}
	err = json.Unmarshal(body, &link)
	if err != nil {
		return "", nil, fmt.Errorf("yandex: invalid %s JSON response: %w", action, err)
	}

	resourceURL, err = url.Parse(link.HREF)
	if err != nil {
		return "", nil, err
	}

	return link.Method, resourceURL, err
}
