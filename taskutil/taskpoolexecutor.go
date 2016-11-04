package taskutil

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	//	log "github.com/alecthomas/log4go"
)

type TaskPoolExecutor struct {
	queueChan     chan *runable
	running       int32 //0(false) or 1(true)
	engines       int
	activeCount   int32
	taskQueueSize int
	PrintPanic    bool
	wg            *sync.WaitGroup
	locker        sync.RWMutex
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
	exetor.wg = &sync.WaitGroup{}
	return exetor
}

func (this *TaskPoolExecutor) GetQueueSize() int {
	return len(this.queueChan)
}

func (this *TaskPoolExecutor) GetActiveCount() int {
	return int(atomic.LoadInt32(&this.activeCount))
}

func (this *TaskPoolExecutor) Start() {
	this.locker.Lock()
	defer this.locker.Unlock()

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
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.isRunning() {
		atomic.StoreInt32(&this.running, 0)
		close(this.queueChan)
	}
}

//关闭并等待所有任务执行完
func (this *TaskPoolExecutor) CloseAndWait() {
	this.Close()
	this.Wait()
}

func (this *TaskPoolExecutor) Wait() {
	this.wg.Wait()
}

func (this *TaskPoolExecutor) isRunning() bool {
	return atomic.LoadInt32(&this.running) > 0
}

//安排任务
func (this *TaskPoolExecutor) Execute(f func(p ...interface{}), p ...interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()
	if !this.isRunning() {
		panic("task pool executor not is running!")
	}
	this.queueChan <- &runable{f: f, ps: p}
}

func (this *TaskPoolExecutor) startEngine(runablech chan *runable) {
	this.wg.Add(1)
	defer this.wg.Done()

	for !(!this.isRunning() && len(runablech) == 0) {
		r, ok := <-runablech
		if ok {
			atomic.AddInt32(&this.activeCount, 1)
			this.executeTask(r)
			atomic.AddInt32(&this.activeCount, -1)
		} else {
			//log.Println("[WRN] runable chan is closed!", r, ok)
			return
		}
	}
}

func (this *TaskPoolExecutor) executeTask(runable *runable) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERR] execute task runtime error caught: %v/n", r)
			if this.PrintPanic {
				for i := 1; ; i += 1 {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Printf("[ERR] %v(%v)\n", file, line)
				}
			}
		}
	}()
	runable.f(runable.ps...)
}
