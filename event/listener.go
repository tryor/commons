package event

type IListener interface {
	HandleEvent(e IEvent) bool //处理事件, 返回false, 将继续处理事件，否则中止事件处理
}
