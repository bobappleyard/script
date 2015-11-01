package script

import (
    "testing"
    "reflect"
)

func TestCodeStep(t *testing.T) {
    p := new(Process)
    p.code = Code{12, 0, 1, 50, 13, 1, 10}
    if p.nextByte() != 12 {
        t.Error("nextByte failed")
    }
    if p.next2Bytes() != 256 {
        t.Error("next2Bytes failed")
    }
    if p.next4Bytes() != 167841074 {
        t.Error("next4Bytes failed")
    }
}

func TestValueStack(t *testing.T) {
    p := new(Process)
    p.push(Int(0))
    p.push(Int(1))
    p.push(Int(2))
    if i, ok := p.pop().AsInt(); !ok || i != 2 {
        t.Error("expecting 2")
    }
    if i, ok := p.pop().AsInt(); !ok || i != 1 {
        t.Error("expecting 1")
    }
    if i, ok := p.pop().AsInt(); !ok || i != 0 {
        t.Error("expecting 0")
    }
}

func TestShuffle(t *testing.T) {
    for i, test := range ([]struct{frame, add, after []V; argc int}{
        {[]V{}, []V{}, []V{}, 0},
        {[]V{Int(1)}, []V{}, []V{Int(1)}, 0},
        {[]V{}, []V{Int(1)}, []V{}, 0},
        {[]V{}, []V{Int(1)}, []V{Int(1)}, 1},
        {[]V{Int(1)}, []V{Int(2)}, []V{Int(1)}, 0},
        {[]V{Int(1)}, []V{Int(2)}, []V{Int(1), Int(2)}, 1},
        {[]V{Int(1)}, []V{Int(2), Int(3)}, []V{Int(1), Int(3)}, 1},
    }) {
        p := new(Process)
        p.stack = test.frame
        p.base = len(test.frame)
        for _, v := range test.add {
            p.push(v)
        }
        p.shuffle(test.argc)
        if !reflect.DeepEqual(p.stack, test.after) {
            t.Errorf("[%d]: %#v != %#v", i, p.stack, test.after)
        }
    }
}

func TestCode1(t *testing.T) {
    for i, test := range ([]struct{unit []V; code Code; result V}{
        {[]V{Int(1)}, Code{4, 0,0,0,0, 0}, Int(1)},
    }) {
        p := new(Process)
        p.unit = &unit{test.unit}
        p.code = test.code
        p.run()
        if p.result != test.result {
            t.Errorf("[%d]: %#v != %#v", i, p.result, test.result)
        }
    }
}





