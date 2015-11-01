package bytecode

import (
    "io"
    "errors"
    "math"
    "encoding/binary"
    
    "fmt"
)

type ItemId uint32
type TypeId byte

// In order to read an image in the bytecode format, implement this interface.
//
// Everything that is read from a bytecode image is given an ItemId so that it
// can be referred to later on in the reading process. The first item has an id
// of 0, the second 1 and so on. This is implicit in the interface and you are
// expected to keep track of this yourself. The most obvious way to implement
// this is to place everything that has been read into an array.
type Handler interface {
    // Atomic items
    Int(x int64) error
    Float(x float64) error
    Bytes(bs []byte) error
    // Compound items
    CompoundSize(id TypeId) (int, error)
    Compound(id TypeId, items []ItemId) error
}


const (
    magicString = "\x00SCR"
    formatString = "\x01\x00\x00\x00"
)

var (
    ErrMagicNumber = errors.New("wrong magic number")
    ErrFormatVersion = errors.New("unsupported format version")
    ErrUnknownSection = errors.New("unknown type id")
    ErrInvalidEntry = errors.New("invalid image item")
    ErrIntTooLarge = errors.New("varint too large (>64 bits)")
)

const (
    Int = iota
    Float
    Bytes
)

func ReadImage(input io.Reader, handler Handler) error {
    return (&reader{input: input, handler: handler}).read()
}

type reader struct {
    input io.Reader
    buf []byte
    handler Handler
    err error
    lastItem ItemId
}

func (r *reader) read() error {
    r.init()
    for r.err == nil {
        r.readItem()
    }
    r.ignoreEOF()
    return r.err
}

func (r *reader) ignoreEOF() {
    if r.err == io.EOF {
        r.err = nil
    }
}

func (r *reader) hasLen(size uint32) bool {
    if len(r.buf) < int(size) {
        r.err = io.ErrUnexpectedEOF
        return false
    }
    return true
}

func (r *reader) readBuffer(size uint32) {
    r.buf = make([]byte, size)
    n, err := io.ReadFull(r.input, r.buf)
    r.err = err
    if n < int(size) {
        r.err = io.ErrUnexpectedEOF
    }
}

func (r *reader) init() {
    size := r.readHead()
    if r.err != nil {
        return
    }
    r.readBuffer(size)
    r.ignoreEOF()
}


func (r *reader) readHead() uint32 {
    r.readBuffer(12)
    eof := r.err == io.EOF
    r.ignoreEOF()
    if r.err != nil {
        return 0
    }
    size := binary.LittleEndian.Uint32(r.buf[8:])
    if eof && size != 0 {
        r.err = io.ErrUnexpectedEOF
    }
    if string(r.buf[4:8]) != formatString {
        r.err = ErrFormatVersion
    }
    if string(r.buf[:4]) != magicString {
        r.err = ErrMagicNumber
    }
    if r.err != nil {
        size = 0
    }
    return size
}

func (r *reader) readItem() {
    if len(r.buf) == 0 {
        r.err = io.EOF
        return
    }
    b := r.nextByte()
    switch b {
    case Int:
        r.readInt()
    case Float:
        r.readFloat()
    case Bytes:
        r.readBytes()
    default:
        r.readCompound(b)
    }
    r.lastItem++
}

func (r *reader) nextByte() byte {
    b := r.buf[0]
    r.buf = r.buf[1:]
    return b
}

func (r *reader) readInt() {
    x, c := binary.Varint(r.buf)
    switch {
    case c > 0:
        r.buf = r.buf[c:]
        r.err = r.handler.Int(x)
    case c < 0:
        r.err = ErrIntTooLarge
    default:
        r.err = io.ErrUnexpectedEOF
    }
}

func (r *reader) readFloat() {
    if !r.hasLen(8) {
        return
    }
    bits := binary.LittleEndian.Uint64(r.buf)
    r.buf = r.buf[8:]
    f := math.Float64frombits(bits)
    r.handler.Float(f)
}

func (r *reader) readSize() (uint32, bool) {
    if !r.hasLen(4) {
        return 0, false
    }
    size := binary.LittleEndian.Uint32(r.buf)
    r.buf = r.buf[4:]
    return size, true
}

func (r *reader) readBytes() {
    size, ok := r.readSize()
    if !ok {
        return
    }
    if !r.hasLen(size) {
        return
    }
    r.err = r.handler.Bytes(r.buf[:size])
    r.buf = r.buf[size:]
}

func (r *reader) readCompound(idb byte) {
    id := TypeId(idb)
    size := r.readCompoundSize(id)
    fmt.Println(size, r.err)
    if r.err != nil {
        return
    }
    if !r.hasLen(4*uint32(size)) {
        return
    }
    items := make([]ItemId, size)
    for i := range items {
        items[i] = ItemId(binary.LittleEndian.Uint32(r.buf[4*i:]))
        if items[i] > r.lastItem {
            r.err = ErrInvalidEntry
            return
        }
    }
    r.buf = r.buf[4*size:]
    r.handler.Compound(id, items)
}

func (r *reader) readCompoundSize(id TypeId) int {
    size, err := r.handler.CompoundSize(id)
    if err != nil {
        r.err = err
        return 0
    }
    if size < 0 {
        c, ok := r.readSize()
        if !ok {
            return 0
        }
        size = int(c)
    }
    return size
}

