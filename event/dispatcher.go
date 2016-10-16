package event

import (
	//"fmt"
	//"sort"
	"sync"
)

type IDispatcher interface {
	RegisterListener(t Type, l IListener, pos ...int)
	RemoveListener(l IListener)
	FireEvent(e IEvent, asyn ...bool) bool
}

type Dispatcher struct {
	//allAsynEvents []*Event //所有异步事件
	//allEventListeners []IListener          //监听所有事件的事件监听器
	AllTypeListeners map[Type][]IListener //具体事件类型相关监听器
	Lock             sync.RWMutex
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{}
	InitDispatcher(d)
	return d
}

func InitDispatcher(d *Dispatcher) {
	//d.allAsynEvents = make([]*Event, 0)
	//d.allEventListeners = make([]IListener, 0)
	d.AllTypeListeners = make(map[Type][]IListener)
}

//注册事件监听器, t为事件类型, 此监听器将监听类型为t的事件。 pos为监听器顺序位置
//如果t为ZERO_TYPE类型，将监听所有事件
func (d *Dispatcher) RegisterListener(t Type, l IListener, pos ...int) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	ls := d.AllTypeListeners[t]
	if ls == nil {
		ls = make([]IListener, 0)
		ls = append(ls, l)
		d.AllTypeListeners[t] = ls
	} else {

		lsSize := len(ls)
		p := lsSize
		if len(pos) > 0 && pos[0] < lsSize {
			p = pos[0]
		}
		if p < lsSize {
			//ls_ := make([]IListener, 0)
			//ls_ = append(ls_, ls[0:p]...)
			//ls_ = append(ls_, l)
			//ls_ = append(ls_, ls[p:]...)
			d.AllTypeListeners[t] = insert(ls, p, l)
		} else {
			ls = append(ls, l)
			d.AllTypeListeners[t] = ls
		}
	}
}

func (d *Dispatcher) RemoveListener(l IListener) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	for key, ls := range d.AllTypeListeners {
		for {
			idx := search(ls, l)
			if idx < len(ls) {
				ls = remove(ls, idx, idx+1)
			} else {
				break
			}
		}
		d.AllTypeListeners[key] = ls
	}
}

/**
 * 触发一个事件, 在触发事件时默认同步执行事件监听器。 asyn为true将异步
 * 处理事件，此时需要调用handleAll方法来处理事件,异步事件返回值无意义，
 * 同步事件返回值为true说明至少有一个监听器响应了事件
 */
func (d *Dispatcher) FireEvent(e IEvent, asyn ...bool) bool {
	if len(asyn) > 0 && asyn[0] {
		go d.handle(e)
		return false
	} else {
		return d.handle(e)
	}
}

func (d *Dispatcher) handle(e IEvent) bool {
	d.Lock.RLock()
	defer d.Lock.RUnlock()

	ls := d.AllTypeListeners[e.GetType()]
	for _, l := range ls {
		if l.HandleEvent(e) {
			return true
		}
	}

	ls = d.AllTypeListeners[ZERO_TYPE]
	for _, l := range ls {
		if l.HandleEvent(e) {
			return true
		}
	}

	return false
}

func insert(slice []IListener, index int, insertion ...IListener) []IListener {
	result := make([]IListener, len(slice)+len(insertion))
	at := copy(result, slice[:index])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[index:])
	return result
}

func remove(slice []IListener, start, end int) []IListener {
	return append(slice[:start], slice[end:]...)
}

func search(ls []IListener, l IListener) int {
	for i, v := range ls {
		if v == l {
			return i
		}
	}
	return len(ls)
}
