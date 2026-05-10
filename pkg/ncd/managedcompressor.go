package ncd

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/bits"
	"os/exec"

	"github.com/google/brotli/go/cbrotli"
	"github.com/klauspost/compress/zstd"
)

/*
Return the valid window size configuration for a compressor, given the target window size determined by the data.
Off-bounds targets return errors.
*/
func ComputeCompressorWindowSize(compressorName string, targetWindowSize int) (int, error) {
	var err error
	w := 0

	switch compressorName {
	case "Brotli":
		// Brotli takes the base 2 logarithm of the window size: it must be a value between 10 and 24
		log2w := bits.TrailingZeros(uint(targetWindowSize))

		if log2w < 10 {
			w = 1 << 10
		} else if log2w > 24 {
			w = 1 << 24
			err = fmt.Errorf("Brotli's maximum sliding window size (%d bytes) cannot fit the required %d bytes", 1<<24, targetWindowSize)
		} else {
			w = 1 << log2w
		}
	case "zstd":
		//  The window size fot zstd  must be in the range 2^10 to 2^29
		if targetWindowSize < zstd.MinWindowSize {
			w = zstd.MinWindowSize
		} else if targetWindowSize > zstd.MaxWindowSize {
			w = 1 << 29
			err = fmt.Errorf("zstd's maximum sliding window size (%d bytes) cannot fit the required %d bytes", zstd.MaxWindowSize, targetWindowSize)
		} else {
			w = targetWindowSize
		}
	case "Gzip":
		// The Gzip window size is fixed to 32 KB
		w = 32 * (1 << 10)
		if targetWindowSize > w {
			err = fmt.Errorf("Gzip sliding window size is fixed to 32 KB cannot fit the required %d bytes", targetWindowSize)
		}
	}

	return w, err
}

/*========================================================================
	INTERFACE
········································································*/

type ManagedCompressor interface {
	// Sends data to the compressor buffer, return the number of bytes sent
	// Multiple calls concatenate data in the buffer
	Send([]byte) (int, error)

	// Returns the compressed size of the data in the compressor buffer and resets the state of the compressor
	Process() int
}

type ManagedCompressorFactory func() ManagedCompressor

/*========================================================================
	BYTE COUNTER
········································································*/

/*
Fake buffer that keeps the count of bytes sent to it.
Implements the io.WriteCloser interface.
*/
type ByteCounter struct {
	nBytes int
}

func (c *ByteCounter) Write(b []byte) (int, error) {
	n := len(b)
	c.nBytes += n

	return n, nil
}

// Does nothing, just to satisfy the io.WriteCloser interface
func (c *ByteCounter) Close() error {
	return nil
}

// Set the counter to zero
func (c *ByteCounter) Reset() {
	c.nBytes = 0
}

/*=======================================================================
	GZIP
·······································································*/

// Wrapper for the Gzip compressor, implements the ManagedCompressor interface
type ManagedCompressorGzip struct {
	compressor *gzip.Writer
	buffer     *ByteCounter
}

func NewManagedCompressorGzip() *ManagedCompressorGzip {
	buffer := &ByteCounter{}
	compressor, _ := gzip.NewWriterLevel(buffer, gzip.BestCompression)

	return &ManagedCompressorGzip{
		compressor: compressor,
		buffer:     buffer,
	}
}

func (mc *ManagedCompressorGzip) Send(data []byte) (int, error) {
	return mc.compressor.Write(data)
}

func (mc *ManagedCompressorGzip) Process() int {
	mc.compressor.Close()
	b := mc.buffer.nBytes
	mc.buffer.Reset()
	mc.compressor.Reset(mc.buffer)

	return b
}

/*
Create a factory of Gzip compressors
*/
func NewManagedCompressorGzipFactory() ManagedCompressorFactory {
	return func() ManagedCompressor { return NewManagedCompressorGzip() }
}

/*=======================================================================
	BROTLI
·······································································*/

// Wrapper for the Brotli compressor, implements the ManagedCompressor interface
type ManagedCompressorBrotli struct {
	buffer     *ByteCounter
	compressor *cbrotli.Writer
	opts       cbrotli.WriterOptions
}

func NewManagedCompressorBrotli(opts cbrotli.WriterOptions) *ManagedCompressorBrotli {
	buffer := &ByteCounter{}

	compressor := cbrotli.NewWriter(buffer, opts)

	return &ManagedCompressorBrotli{buffer: buffer, compressor: compressor, opts: opts}
}

func (mc *ManagedCompressorBrotli) Send(data []byte) (int, error) {
	return mc.compressor.Write(data)
}

func (mc *ManagedCompressorBrotli) Process() int {
	mc.compressor.Close()
	b := mc.buffer.nBytes
	mc.buffer.Reset()

	mc.compressor = cbrotli.NewWriter(mc.buffer, mc.opts)

	return b
}

/*
Create a factory of Brotli compressors for a given window size (in bytes)
*/
func NewManagedCompressorBrotliFactory(windowSize int) ManagedCompressorFactory {
	// The WriterOptions take the log2 of the window size
	log2w := bits.TrailingZeros(uint(windowSize))

	opts := cbrotli.WriterOptions{
		Quality: 11,    // Compression level
		LGWin:   log2w, // Sliding window
	}

	return func() ManagedCompressor { return NewManagedCompressorBrotli(opts) }
}

/*=======================================================================
	zstd
·······································································*/

type ManagedCompressorZstd struct {
	compressor *zstd.Encoder
	buffer     *ByteCounter
	opts       []zstd.EOption
}

func NewManagedCompressorZstd(opts []zstd.EOption) *ManagedCompressorZstd {
	buffer := &ByteCounter{}

	compressor, err := zstd.NewWriter(buffer, opts...)
	if err != nil {
		panic("could not create zstd encoder")
	}
	mc := &ManagedCompressorZstd{compressor: compressor, buffer: buffer, opts: opts}

	return mc
}

func (mc *ManagedCompressorZstd) Send(data []byte) (int, error) {
	return mc.compressor.Write(data)
}

func (mc *ManagedCompressorZstd) Process() int {
	mc.compressor.Close()
	b := mc.buffer.nBytes
	mc.buffer.Reset()

	mc.compressor.Reset(mc.buffer)

	return b
}

/*
Create a factory of zstd compressors for a given window size (in bytes)
*/
func NewManagedCompressorZstdFactory(windowSize int) ManagedCompressorFactory {
	opts := []zstd.EOption{zstd.WithWindowSize(windowSize)}

	return func() ManagedCompressor { return NewManagedCompressorZstd(opts) }
}

/*=======================================================================
	External compressor
·······································································*/

type ManagedCompressorExternal struct {
	args      []string
	stdin     io.WriteCloser
	outbuffer *ByteCounter
	cmd       *exec.Cmd
}

func NewManagedCompressorExternal(args []string) *ManagedCompressorExternal {
	mc := &ManagedCompressorExternal{
		args:      args,
		outbuffer: &ByteCounter{},
	}

	mc.start()
	return mc
}

func (mc *ManagedCompressorExternal) start() {
	mc.cmd = exec.Command(mc.args[0], mc.args[1:]...)
	mc.cmd.Stdout = mc.outbuffer
	stdin, err := mc.cmd.StdinPipe()
	if err != nil {
		panic("could not create external compressor stdin pipe")
	}
	mc.stdin = stdin

	if err := mc.cmd.Start(); err != nil {
		panic("could not start external compressor")
	}
}

func (mc *ManagedCompressorExternal) Send(data []byte) (int, error) {
	return mc.stdin.Write(data)
}

func (mc *ManagedCompressorExternal) Process() int {
	mc.stdin.Close()
	mc.cmd.Wait()
	b := mc.outbuffer.nBytes
	mc.outbuffer.Reset()

	mc.start()

	return b
}

func NewManagedCompressorExternalFactory(args []string) ManagedCompressorFactory {
	return func() ManagedCompressor { return NewManagedCompressorExternal(args) }
}
