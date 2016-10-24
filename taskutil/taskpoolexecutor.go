package taskutil

import (
	"runtime"
	"sync/atomic"

	log "github.com/alecthomas/log4go"
)

type TaskPoolExecutor struct {
	queueChan     chan *runable
	running       int32 //0(false) or 1(true)
	engines       int
	activeCount   int32
	taskQueueSize int
	PrintPanic    bool
}

type runable struct {
	f  func(p ...interface{}) //函数
	ps []interface{}          //参数
}

func NewTaskPoolExecutor(engines int, taskQueueSize int) *TaskPoolExecutor {
	if engines <= 0 {
		engines = runtime.NumCPU() / 2
		if engines < 3 {
			engines = 3
		}
	}

	exetor := &TaskPoolExecutor{running: 0, engines: engines, taskQueueSize: taskQueueSize, PrintPanic: true}
	return exetor
}

func (this *TaskPoolExecutor) GetQueueSize() int {
	return len(this.queueChan)
}

func (this *TaskPoolExecutor) GetActiveCount() int {
	return int(atomic.LoadInt32(&this.activeCount))
}

func (this *TaskPoolExecutor) Start() {
	if this.isRunning() {
		return
	}
	this.running = 1
	this.queueChan = make(chan *runable, this.taskQueueSize)
	for i := 0; i < this.engines; i++ {
		go this.startEngine(this.queueChan)
	}
}

func (this *TaskPoolExecutor) Close() {
	atomic.StoreInt32(&this.running, 0)
}

func (this *TaskPoolExecutor) isRunning() bool {
	return atomic.LoadInt32(&this.running) > 0
}

//安排任务
func (this *TaskPoolExecutor) Execute(f func(p ...interface{}), p ...interface{}) {
	if !this.isRunning() {
		panic("task pool executor not is running!")
	}
	this.queueChan <- &runable{f: f, ps: p}
}

func (this *TaskPoolExecutor) startEngine(runablech chan *runable) {
	for !(!this.isRunning() && len(runablech) == 0) {
		r, ok := <-runablech
		if ok {
			atomic.AddInt32(&this.activeCount, 1)
			this.executeTask(r)
			atomic.AddInt32(&this.activeCount, -1)
		} else {
			log.Warn("runable chan is closed!")
			return
		}
	}
	//fmt.Printf("this.isRunning():%v, len(runablech):%v\n", this.isRunning(), len(runablech))
	close(runablech)
}

func (this *TaskPoolExecutor) executeTask(runable *runable) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("execute task runtime error caught: %v", r)
			if this.PrintPanic {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Info("%v %v", file, line)
				}
			}
		}
	}()
	runable.f(runable.ps...)
}
