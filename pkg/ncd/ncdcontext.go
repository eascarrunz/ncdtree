package ncd

import (
	"compress/gzip"
	"io"

	"github.com/google/brotli/go/cbrotli"
)

type genericCompressor interface {
	Write([]byte) (int, error)
	Close() error
	Reset(io.Writer)
}

type ByteCounter struct {
	nBytes int
}

func (c *ByteCounter) Write(b []byte) (int, error) {
	n := len(b)
	c.nBytes += n

	return n, nil
}

func (c *ByteCounter) Close() error {
	return nil
}

func (c *ByteCounter) Reset() {
	c.nBytes = 0
}

type CompressionContext struct {
	Compressor genericCompressor
	Counter    *ByteCounter
}

func NewGzipCompressionContext() *CompressionContext {
	counter := &ByteCounter{}
	compressor := gzip.NewWriter(counter)

	return &CompressionContext{
		Compressor: compressor,
		Counter:    counter,
	}
}

type BrotliWriterReset struct {
	w    *cbrotli.Writer
	opts cbrotli.WriterOptions
}

func NewBrotliWriterReset(buf io.Writer, opts cbrotli.WriterOptions) *BrotliWriterReset {
	w := cbrotli.NewWriter(buf, opts)

	return &BrotliWriterReset{w, opts}
}

func (w *BrotliWriterReset) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *BrotliWriterReset) Close() error {
	return w.w.Close()
}

func (w *BrotliWriterReset) Reset(ww io.Writer) {
	w.w = cbrotli.NewWriter(ww, w.opts)
}

func NewBrotliCompressionContext() *CompressionContext {
	counter := &ByteCounter{}
	compressor_options := cbrotli.WriterOptions{
		Quality: 11,
		LGWin:   0,
	}
	compressor := NewBrotliWriterReset(counter, compressor_options)

	return &CompressionContext{
		Compressor: compressor,
		Counter:    counter,
	}
}

// Return the number of compressed bytes and reset the compressor and counter
func (ctx *CompressionContext) SizeReset() int {
	ctx.Compressor.Close()
	b := ctx.Counter.nBytes
	ctx.Counter.Reset()
	ctx.Compressor.Reset(ctx.Counter)

	return b
}

func (ctx *CompressionContext) Write(b []byte) {
	ctx.Compressor.Write(b)
}
