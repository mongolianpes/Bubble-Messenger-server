package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func CompressGzip(data []byte) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(buf)

	_, err := gz.Write(data)
	if err != nil {
		gzipWriterPool.Put(gz)
		bufferPool.Put(buf)
		return nil, err
	}

	if err := gz.Close(); err != nil {
		gzipWriterPool.Put(gz)
		bufferPool.Put(buf)
		return nil, err
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())

	gzipWriterPool.Put(gz)
	bufferPool.Put(buf)

	return out, nil
}

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func DecompressGzip(compressed []byte) ([]byte, error) {
	r := gzipReaderPool.Get().(*gzip.Reader)
	if err := r.Reset(bytes.NewReader(compressed)); err != nil {
		gzipReaderPool.Put(r)
		return nil, err
	}

	data, err := io.ReadAll(r)
	r.Close()
	gzipReaderPool.Put(r)

	return data, err
}
