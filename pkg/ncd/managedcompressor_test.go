package ncd

import (
    "bytes"
    "testing"

    "github.com/google/brotli/go/cbrotli"
)

func TestByteCounter_Write_Reset(t *testing.T) {
    tests := []struct {
        name   string
        inputs [][]byte
        want   int
    }{
        {"empty", [][]byte{}, 0},
        {"single small", [][]byte{[]byte("abc")}, 3},
        {"multiple small", [][]byte{[]byte("a"), []byte("bc"), []byte("d")}, 4},
        {"large", [][]byte{bytes.Repeat([]byte("x"), 1000)}, 1000},
        {"mixed", [][]byte{[]byte("hello"), []byte(""), []byte("world")}, 10},
        {"unicode", [][]byte{[]byte("你好"), []byte("世界")}, 12}, // each Chinese char is 3 bytes in UTF-8
    }

    for _, tt := range tests {
        c := &ByteCounter{}
        for _, in := range tt.inputs {
            n, err := c.Write(in)
            if n != len(in) || err != nil {
                t.Errorf("%s: Write(%q) got (%d,%v), want (%d,nil)", tt.name, in, n, err, len(in))
            }
        }
        if c.nBytes != tt.want {
            t.Errorf("%s: nBytes = %d, want %d", tt.name, c.nBytes, tt.want)
        }
        c.Reset()
        if c.nBytes != 0 {
            t.Errorf("%s: Reset did not zero nBytes", tt.name)
        }
    }
}

func TestManagedCompressorGzip_Send_Process(t *testing.T) {
    tests := []struct {
        name   string
        inputs [][]byte
    }{
        {"empty", [][]byte{}},
        {"single", [][]byte{[]byte("abc")}},
        {"multiple", [][]byte{[]byte("a"), []byte("bc"), []byte("d")}},
        {"large", [][]byte{bytes.Repeat([]byte("x"), 1000)}},
        {"unicode", [][]byte{[]byte("你好世界")}},
        {"mixed", [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}},
    }

    for _, tt := range tests {
        mc := NewManagedCompressorGzip()
        for _, in := range tt.inputs {
            _, err := mc.Send(in)
            if err != nil {
                t.Errorf("%s: Send(%q) error: %v", tt.name, in, err)
            }
        }
        size := mc.Process()
        if size <= 0 && len(tt.inputs) > 0 {
            t.Errorf("%s: Process() size = %d, want > 0", tt.name, size)
        }
    }
}

func TestManagedCompressorBrotli_Send_Process(t *testing.T) {
    tests := []struct {
        name   string
        inputs [][]byte
    }{
        {"empty", [][]byte{}},
        {"single", [][]byte{[]byte("abc")}},
        {"multiple", [][]byte{[]byte("a"), []byte("bc"), []byte("d")}},
        {"large", [][]byte{bytes.Repeat([]byte("x"), 1000)}},
        {"unicode", [][]byte{[]byte("你好世界")}},
        {"mixed", [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}},
    }

    opts := cbrotli.WriterOptions{Quality: 5, LGWin: 0}
    for _, tt := range tests {
        mc := NewManagedCompressorBrotli(opts)
        for _, in := range tt.inputs {
            _, err := mc.Send(in)
            if err != nil {
                t.Errorf("%s: Send(%q) error: %v", tt.name, in, err)
            }
        }
        size := mc.Process()
        if size < 0 {
            t.Errorf("%s: Process() size = %d, want >= 0", tt.name, size)
        }
    }
}