package yandex

import (
	"io/fs"
	"time"
)

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}

func (fi *fileInfo) Mode() fs.FileMode {
	return 0
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *fileInfo) IsDir() bool {
	return false
}

func (fi *fileInfo) Sys() interface{} {
	return nil
}
