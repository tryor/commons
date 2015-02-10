package event

type Type int

const ZERO_TYPE Type = 0

type IEvent interface {
	GetType() Type
	GetSource() interface{}
}

type Event struct {
	Type   Type        //事件类型
	Source interface{} //事件源
}

func NewEvent(t Type, source interface{}) *Event {
	return &Event{Type: t, Source: source}
}

func (e *Event) GetType() Type {
	return e.Type
}

func (e *Event) GetSource() interface{} {
	return e.Source
}
