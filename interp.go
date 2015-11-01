package script

import ()

type Interpreter struct {
    builtins builtins
    packageRoot V
    names map[string]*Name
}

type Code []byte

type unit struct {
    Values []V
}

type Process struct {
    host *Interpreter
    result V
    frame
    control []frame
}

type frame struct {
    this, slot, handler V
    argc, pos, base int
    code Code
    closure []V
    stack []V
    unit *unit
}

func New() *Interpreter {
    return new(Interpreter).init()
}

func (host *Interpreter) init() *Interpreter {
    return host
}

const (
    HALT = iota
    THIS
    BOUND
    FREE
    GLOBAL
    JUMP
    BRANCH
    PUSH
    LOOKUP
    GET
    SET
    CALL
    TGET
    TCALL
    FRAME
    RETURN
)

func (p *Process) run() {
    for {
        switch p.nextByte() {
        case HALT:
            return
        case THIS:
            p.result = p.this
        case BOUND:
            n := p.nextByte()
            p.result = p.stack[p.base + n]
        case FREE:
            n := p.next2Bytes()
            p.result = p.stack[n]
        case GLOBAL:
            id := p.next4Bytes()
            p.result = p.unit.Values[id]
        case JUMP:
            loc := p.next2Bytes()
            p.pos = loc
        case BRANCH:
            bpos := p.next2Bytes()
            if !p.result.AsBool() {
                p.pos = bpos
            }
        case PUSH:
            p.push(p.result)
        case LOOKUP:
            id := p.next4Bytes()
            name := p.unit.Values[id]
            p.lookup(name)
        case GET:
            p.get(false)
        case SET:
            val := p.pop()
            p.result = V{nil}
            p.set(val)
        case CALL:
            argc := p.nextByte()
            p.call(argc, false)
        case TGET:
            p.get(true)
        case TCALL:
            argc := p.nextByte()
            p.call(argc, true)
        case FRAME:
            loc := p.next2Bytes()
            p.pos = loc
            p.enter()
        case RETURN:
            p.leave()
        }
    }
}

func (p *Process) nextByte() int {
    res := p.code[p.pos]
    p.pos++
    return int(res)
}

func (p *Process) next2Bytes() int {
    b1 := p.nextByte()
    b2 := p.nextByte()
    return (b2 << 8) + b1
}

func (p *Process) next4Bytes() int {
    b1 := p.next2Bytes()
    b2 := p.next2Bytes()
    return (b2 << 16) + b1
}

func (p *Process) push(x V) {
    p.stack = append(p.stack, x)
}

func (p *Process) pop() V {
    end := len(p.stack)-1
    last := p.stack[end]
    p.stack = p.stack[:end]
    return last
}

func (p *Process) enter() {
    p.control = append(p.control, p.frame)
}

func (p *Process) leave() {
    end := len(p.control)-1
    p.frame = p.control[end]
    p.control = p.control[:end]
}

func (p *Process) lookup(nm V) {
    nmv, ok := nm.val.(*Name)
    if !ok {
        p.throw("name wrong type")
        return
    }
    cls := p.host.ClassOf(p.result)
    bcls, ok := cls.val.(*class)
    if ok {
        slot, err := bcls.lookup(nmv)
        if err != nil {
            p.throw(err.Error())
        }
        p.slot = slot
        return
    }
    p.push(p.result)
    p.enter()
    p.push(nm)
    p.result = cls
    p.lookup(p.host.builtins.names.lookup)
    p.slot = p.result
    p.result = p.pop()
}

func (p *Process) getFieldOffset() (*V, bool) {
    offset, ok := p.slot.val.(*UserObject).fields[0].AsInt()
    if !ok {
        p.throw("unexpected field offset type")
        return nil, false
    }
    obj, ok := p.this.AsObject()
    if !ok {
        p.throw("unexpected target type")
        return nil, false
    }
    return &obj.fields[offset], true
}

func (p *Process) get(tail bool) {
    if p.host.ClassOf(p.slot) == p.host.builtins.classes.Field {
        if field, ok := p.getFieldOffset(); ok {
            p.result = *field
        }
    }
    if !tail {
        p.enter()
    }
    p.push(p.result)
    p.lookup(p.host.builtins.names.getSlot)
    p.call(1, tail)
}

func (p *Process) set(val V) {
    if p.host.ClassOf(p.slot) == p.host.builtins.classes.Field {
        if field, ok := p.getFieldOffset(); ok {
            *field = val
        }
    }
    p.enter()
    p.push(val)
    p.push(p.result)
    p.lookup(p.host.builtins.names.setSlot)
    p.call(2, false)
}

func (p *Process) call(argc int, tail bool) {
    calln := p.host.builtins.names.callSlot
    for !p.callPrimitive(argc, tail) {
        p.push(p.result)
        p.result = p.slot
        argc++
        p.lookup(calln)
    }
}

func (p *Process) callPrimitive(argc int, tail bool) bool {
    if p.host.ClassOf(p.slot) != p.host.builtins.classes.Primitive {
        return false
    }
    if tail {
        p.shuffle(argc)
    }
    fn, ok := p.result.val.(Primitive)
    if !ok {
        p.throw("unexpected primitive method type")
        return true
    }
    p.argc = argc
    fn(p).perform(p)
    return true
}

func (p *Process) shuffle(argc int) {
    dest := p.base
    src := len(p.stack) - argc
    for i := 0; i < argc; i++ {
        p.stack[dest+i] = p.stack[src+i]
    }
    p.stack = p.stack[:dest+argc]
}

func (p *Process) throw(e string) {
    panic(e)
}


