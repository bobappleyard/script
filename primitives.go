package script

import ()

type Primitive func(*Process) Action

type Action struct {
    kind int
    data V
}

type builtins struct {
    classes builtinClasses
    names builtinNames
}


func (a Action) perform(p *Process) {
}

type builtinNames struct {
    lookup, getSlot, setSlot, callSlot V
}



