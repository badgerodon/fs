package fs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
)

var registry = struct {
	sync.RWMutex
	m map[string]func() FileSystem
}{
	m: map[string]func() FileSystem{},
}

// Register registers a filesystem provider for the given scheme.
func Register(scheme string, constructor func() FileSystem) {
	registry.Lock()
	registry.m[scheme] = constructor
	registry.Unlock()
}

// Copy copies a file from srcURL to dstURL.
func Copy(ctx context.Context, dstURL, srcURL *url.URL) error {
	src, err := Open(ctx, srcURL)
	if err != nil {
		return fmt.Errorf("error opening source: %w", err)
	}

	dst, err := Create(ctx, dstURL)
	if err != nil {
		_ = src.Close()
		return fmt.Errorf("error opening destination: %w", err)
	}

	_, err = io.Copy(dst, NewProgressReader(src, GetProgressFunc(ctx)))
	if err != nil {
		_ = src.Close()
		_ = dst.Close()
		return fmt.Errorf("error copying from source to destination: %w", err)
	}

	err = src.Close()
	if err != nil {
		_ = dst.Close()
		return fmt.Errorf("error closing source: %w", err)
	}

	err = dst.Close()
	if err != nil {
		return fmt.Errorf("error closing destination: %w", err)
	}

	return nil
}

// Create creates a file.
func Create(ctx context.Context, dstURL *url.URL) (io.WriteCloser, error) {
	registry.RLock()
	dstFS, ok := registry.m[dstURL.Scheme]
	registry.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown destination scheme: %s", dstURL.Scheme)
	}

	return dstFS().Create(ctx, dstURL)
}

// Open opens a file.
func Open(ctx context.Context, srcURL *url.URL) (io.ReadCloser, error) {
	registry.RLock()
	srcFS, ok := registry.m[srcURL.Scheme]
	registry.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown source scheme: %s", srcURL.Scheme)
	}
	return srcFS().Open(ctx, srcURL)
}

// Stat gets the file info for a file.
func Stat(ctx context.Context, srcURL *url.URL) (FileInfo, error) {
	registry.RLock()
	srcFS, ok := registry.m[srcURL.Scheme]
	registry.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown source scheme: %s", srcURL.Scheme)
	}
	return srcFS().Stat(ctx, srcURL)
}
