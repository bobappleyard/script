package bytecode

import (
    "testing"
    "strings"
    "reflect"
    "io"
//    "fmt"
)

var testTypes = []struct{name string; size int}{
    {"test1", 1},
    {"test2", 0},
    {"test3", -1},
}

type testHandler struct {
    items []testItem
}

type testItem struct {
    kind string
    value interface{}
}

func (h *testHandler) addItem(kind string, value interface{}) {
    h.items = append(h.items, testItem{kind, value})
}

func (h *testHandler) Int(x int64) error {
    h.addItem("int", x)
    return nil
}

func (h *testHandler) Float(x float64) error {
    h.addItem("float", x)
    return nil
}

func (h *testHandler) Bytes(x []byte) error {
    h.addItem("bytes", x)
    return nil
}

func (h *testHandler) CompoundSize(id TypeId) (int, error) {
    idx := int(id)-3
    if idx >= len(testTypes) {
        return 0, ErrUnknownSection
    }
    return testTypes[idx].size, nil
}

func (h *testHandler) Compound(id TypeId, items []ItemId) error {
    h.addItem(testTypes[int(id)-3].name, items)
    return nil
}

func TestRead(t *testing.T) {
    header := "\x00SCR\x01\x00\x00\x00"
    for i, test := range ([]struct{inp string; err error; items []testItem}{
        {header, io.ErrUnexpectedEOF, nil},
        {header + "\x00\x00\x00\x00", nil, nil},
        {header + "\x09\x00\x00\x00\x00\x00", io.ErrUnexpectedEOF, nil},
        {header + "\x02\x00\x00\x00\x00\x00", nil, []testItem{{"int", int64(0)}}},
        {header + "\x09\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00", nil, []testItem{{"float", float64(0)}}},
        {header + "\x09\x00\x00\x00\x02\x04\x00\x00\x00\x01\x02\x03\x04", nil, []testItem{{"bytes", []byte{1,2,3,4}}}},
        {header + "\x02\x00\x00\x00\x00\x00", nil, []testItem{{"int", int64(0)}}},
        {header + "\x07\x00\x00\x00\x00\x00\x03\x00\x00\x00\x00", nil, []testItem{{"int", int64(0)}, {"test1", []ItemId{0}}}},
        {header + "\x01\x00\x00\x00\x04", nil, []testItem{{"test2", []ItemId{}}}},
        {header + "\x05\x00\x00\x00\x05\x00\x00\x00\x00", nil, []testItem{{"test3", []ItemId{}}}},
    }) {
        handler := &testHandler{}
        err := ReadImage(strings.NewReader(test.inp), handler)
        if err != test.err {
            t.Errorf("[%d] unexpected error (expected: %s, got: %s)", i, test.err, err)
        }
        if !reflect.DeepEqual(handler.items, test.items) {
            t.Errorf("[%d] unexpected items (expected: %#v, got: %#v)", i, test.items, handler.items)
        }
    }
}

