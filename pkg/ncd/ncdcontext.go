package ncd

import (
	"compress/gzip"
	"io"

	"github.com/google/brotli/go/cbrotli"
	"github.com/ulikunitz/xz/lzma"
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

type BrotliWriterReseter struct {
	w    *cbrotli.Writer
	opts cbrotli.WriterOptions
}

func NewBrotliWriterReset(buf io.Writer, opts cbrotli.WriterOptions) *BrotliWriterReseter {
	w := cbrotli.NewWriter(buf, opts)

	return &BrotliWriterReseter{w, opts}
}

func (w *BrotliWriterReseter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *BrotliWriterReseter) Close() error {
	return w.w.Close()
}

func (w *BrotliWriterReseter) Reset(ww io.Writer) {
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

type LZMA2WriteReseter struct {
	writer *lzma.Writer2
}

func NewLZMA2WriteReseter(buf io.Writer) *LZMA2WriteReseter {
	writer, err := lzma.NewWriter2(buf)
	if err != nil {
		panic(err)
	}

	return &LZMA2WriteReseter{writer}
}

func (wr LZMA2WriteReseter) Write(b []byte) (int, error) {
	return wr.writer.Write(b)
}

func (wr LZMA2WriteReseter) Close() error {
	return wr.writer.Close()
}

func (wr LZMA2WriteReseter) Reset(buf io.Writer) {
	writer, err := lzma.NewWriter2(buf)
	if err != nil {
		panic(err)
	}
	wr.writer = writer
}

func NewLZMACompressionContext() *CompressionContext {
	counter := &ByteCounter{}
	// compressorOptions := lzma.Writer2Config
	compressor := NewLZMA2WriteReseter(counter)

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
