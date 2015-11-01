package script

import (
    "sync/atomic"
    "unsafe"
    "sort"
)

// A value in the scripting language. Everything accessible from user code is an
// instance of this type.
type V struct {
    val interface{}
}

// An object defined within the runtime, i.e. not a primitive.
type UserObject struct {
    class V
    fields []V
}

// All values have a type. This is usually an instance of a class.
type class struct {
    ancestor *class
    shape *shape
    names []*Name
    values []V
}

type entityId uint32

// Represents the name of an object member.
type Name struct {
    id entityId
    str string
    parent *Name
    __items *[]nameItem
}

// All the root types the owning name appears in and where.
type nameItem struct {
    introduced entityId
    offset int
}

// Classes have shapes. Usually the relationship is 1:1 but in some circumstances,
// e.g. where a function creates and returns a class, multiple classes may share
// a shape.
type shape struct {
    id entityId
    shapeset []entityId
    names []*Name
    size int
    __children *[]*shape
}

func (c *class) lookup(n *Name) (res V, err error) {
    idx := c.shape.lookup(n)
    if idx == -1 {
        return
    }
    res = c.values[idx]
    return
}

type atomicCounter uint32

func (c *atomicCounter) next() entityId {
    return entityId(atomic.AddUint32((*uint32)(c), 1))
}

var shapeIds = new(atomicCounter)
var nameIds = new(atomicCounter)

func (n *Name) init(str string) *Name {
    n.id = nameIds.next()
    n.str = str
    n.__items = &[]nameItem{}
    return n
}

func (n *Name) getItemsLoc() *unsafe.Pointer {
    return (*unsafe.Pointer)(unsafe.Pointer(&n.__items))
}

func (n *Name) getItems() []nameItem {
    p := atomic.LoadPointer(n.getItemsLoc())
    return *(*[]nameItem)(p)
}

type itemSorter []nameItem
func (s itemSorter) Len() int { return len(s) }
func (s itemSorter) Less(n, m int) bool { return s[n].introduced < s[m].introduced }
func (s itemSorter) Swap(n, m int) { s[n], s[m] = s[m], s[n] }

func (n *Name) appendItem(x nameItem) {
retry:
    p := n.getItemsLoc()
    oldP := atomic.LoadPointer(p)
    oldItems := *(*[]nameItem)(oldP)
    items := make([]nameItem, len(oldItems)+1)
    copy(items, oldItems)
    items[len(oldItems)] = x
    sort.Sort(itemSorter(items))
    newP := unsafe.Pointer(&items)
    if !atomic.CompareAndSwapPointer(p, oldP, newP) {
        goto retry
    }
}

func (s *shape) init(shapeset []entityId, names []*Name, size int) *shape {
    s.id = shapeIds.next()
    ids := make([]entityId, len(shapeset) + 1)
    copy(ids, shapeset)
    ids[len(shapeset)] = s.id
    s.shapeset = ids
    s.names = names
    s.size = size
    s.__children = &[]*shape{}
    return s
}

func (s *shape) getChildrenLoc() *unsafe.Pointer {
    return (*unsafe.Pointer)(unsafe.Pointer(&s.__children))
}

func (s *shape) getChildren() []*shape {
    p := atomic.LoadPointer(s.getChildrenLoc())
    return *(*[]*shape)(p)
}

func (s *shape) tryAppendChild(x *shape) bool {
    p := s.getChildrenLoc()
    oldP := atomic.LoadPointer(p)
    oldChildren := *(*[]*shape)(oldP)
    children := make([]*shape, len(oldChildren)+1)
    copy(children, oldChildren)
    children[len(oldChildren)] = x
    newP := unsafe.Pointer(&children)
    return !atomic.CompareAndSwapPointer(p, oldP, newP)
}

func (s *shape) lookup(n *Name) int {
    items := n.getItems()
    nLen, sLen := len(items), len(s.shapeset)
    // This assumes that the two arrays are sorted by shape ID
    for nCur, sCur := 0, 0; nCur < nLen && sCur < sLen; {
        nameItem := items[nCur]
        shapeId := s.shapeset[sCur]
        if shapeId == nameItem.introduced {
            return nameItem.offset
        } else if shapeId > nameItem.introduced {
            nCur++
        } else {
            sCur++
        }
    }
    return -1
}

func (s *shape) extend(ns []*Name) *shape {
    var missing []*Name
    for _, n := range ns {
        if s.lookup(n) == -1 {
            missing = append(missing, n)
        }
    }
    if len(missing) == 0 {
        return s
    }
retry:
    children := s.getChildren()
    for _, child := range children {
        if child.contains(missing) {
            return child
        }
    }
    // Create a new child
    child := new(shape).init(s.shapeset, missing, s.size + len(missing))
    // Inform the names about the new shape
    for i, n := range missing {
        n.appendItem(nameItem{child.id, s.size + i})
    }
    // A similar shape may have been added while we were working
    if !s.tryAppendChild(child) {
        goto retry
    }
    return child
}

func (s *shape) contains(ns []*Name) bool {
    for _, n := range ns {
        if s.lookup(n) == -1 {
            return false
        }
    }
    return true
}

// Basic objects

func Int(x int64) V {
    return V{x}
}

func Float(x float64) V {
    return V{x}
}

func String(x string) V {
    return V{x}
}

func (v V) AsInt() (val int64, ok bool) {
    val, ok = v.val.(int64)
    return
}

func (v V) AsFloat() (val float64, ok bool) {
    val, ok = v.val.(float64)
    return
}

func (v V) AsString() (val string, ok bool) {
    val, ok = v.val.(string)
    return
}

func (v V) AsObject() (val *UserObject, ok bool) {
    val, ok = v.val.(*UserObject)
    return
}

func (v V) AsBool() bool {
    val, ok := v.val.(bool)
    if !ok {
        return true
    }
    return val
}

func (e *Interpreter) ClassOf(x V) V {
    cs := e.builtins.classes
    switch xv := x.val.(type) {
    case *UserObject:
        return xv.class
    case int64:
        return cs.Integer
    case float64:
        return cs.Float
    case string:
        return cs.String
    case *class:
        return cs.Class
    case *[]V:
        return cs.Array
    case Primitive:
        return cs.Primitive
    }
    panic("unkown object type")
}






