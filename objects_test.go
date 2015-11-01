package script

import (
    "testing"
    "reflect"
)

func TestName(t *testing.T) {
    n := new(Name).init("boom")
    itemsMustBe := func(items []nameItem) {
        itemsAre := n.getItems()
        if !reflect.DeepEqual(itemsAre, items) {
            t.Errorf("failed items test: %#v != %#v", itemsAre, items)
        }
    }
    itemsMustBe([]nameItem{})
    n.appendItem(nameItem{1,1})
    t.Log(n.__items)
    itemsMustBe([]nameItem{{1,1}})
}

func TestShapeExtend(t *testing.T) {
    s := new(shape).init(nil, nil, 0)
    n := new(Name).init("test")
    s2 := s.extend([]*Name{n})
    s3 := s.extend([]*Name{n})
    if s3 != s2 {
        t.Error("extend is not idempotent when introducing new names")
    }
    s4 := s2.extend([]*Name{n})
    if s4 != s2 {
        t.Error("extend is not idempotent with names that have already been introduced")
    }
}

func TestShapeLookup(t *testing.T) {
    s := new(shape).init(nil, nil, 0)
    n1 := new(Name).init("hello")
    n2 := new(Name).init("goodbye")
    s2 := s.extend([]*Name{n1})
    s3 := s.extend([]*Name{n2})
    s4 := s.extend([]*Name{n1, n2})
    
    for i, test := range ([]struct{s *shape; n *Name; there bool}{
        {s, n1, false},
        {s, n2, false},
        {s2, n1, true},
        {s2, n2, false},
        {s3, n1, false},
        {s3, n2, true},
        {s4, n1, true},
        {s4, n2, true},
    }) {
        if test.s.lookup(test.n) != -1 != test.there {
            if test.there {
                t.Errorf("[%d]: %s expected and not found", i, test.n.str)
            } else {
                t.Errorf("[%d]: %s found but not expected", i, test.n.str)
            }
        }
    }
}

func TestBasic(t *testing.T) {
    if val, ok := Int(1).AsInt(); !ok || val != 1 {
        t.Error("Int/AsInt")
    }
    if val, ok := Float(0.1).AsFloat(); !ok || val != 0.1 {
        t.Error("Float/AsFloat")
    }
    if val, ok := String("1").AsString(); !ok || val != "1" {
        t.Error("String/AsString")
    }
    if _, ok := String("hello").AsInt(); ok {
        t.Error("unexpected Int")
    }
    if _, ok := String("hello").AsFloat(); ok {
        t.Error("unexpected Float")
    }
    if _, ok := Int(1).AsString(); ok {
        t.Error("unexpected String")
    }
    if _, ok := Int(1).AsFloat(); ok {
        t.Error("unexpected Float")
    }
    if _, ok := Float(1).AsString(); ok {
        t.Error("unexpected String")
    }
    if _, ok := Float(1).AsInt(); ok {
        t.Error("unexpected Int")
    }
}



