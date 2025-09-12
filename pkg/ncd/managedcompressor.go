package ncd

import (
	"compress/gzip"

	"github.com/google/brotli/go/cbrotli"
)

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
	compressor := gzip.NewWriter(buffer)

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
