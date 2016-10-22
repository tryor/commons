package taskutil

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"

	log "github.com/alecthomas/log4go"
)

type TaskPoolExecutor struct {
	runablechs            []chan *runable
	running               int32 //0(false) or 1(true)
	engineNums            int
	activeTaskNums        int32
	taskQueueSize         int
	randomExecuteTaskMode bool
	engineIdx             int
	lock                  *sync.Mutex
	PrintPanic            bool
}

type runable struct {
	f  func(p ...interface{}) //函数
	ps []interface{}          //参数
}

func NewTaskPoolExecutor(engineNums int, taskQueueSize int, randomExecuteTaskMode ...bool) *TaskPoolExecutor {
	if engineNums <= 0 {
		engineNums = runtime.NumCPU() / 2
		if engineNums < 3 {
			engineNums = 3
		}
	}
	exetor := &TaskPoolExecutor{running: 0, engineNums: engineNums, taskQueueSize: taskQueueSize, PrintPanic: true}
	exetor.runablechs = make([]chan *runable, engineNums)
	exetor.lock = new(sync.Mutex)
	if len(randomExecuteTaskMode) > 0 {
		exetor.randomExecuteTaskMode = randomExecuteTaskMode[0]
	}
	return exetor
}

func (this *TaskPoolExecutor) GetActiveTaskNums() int32 {
	return atomic.LoadInt32(&this.activeTaskNums)
}

func (this *TaskPoolExecutor) incActiveTaskNums() {
	atomic.AddInt32(&this.activeTaskNums, 1)
}

func (this *TaskPoolExecutor) decActiveTaskNums() {
	atomic.AddInt32(&this.activeTaskNums, -1)
}

func (this *TaskPoolExecutor) Start() {
	if this.isRunning() {
		return
	}
	this.running = 1
	for i := 0; i < this.engineNums; i++ {
		this.runablechs[i] = make(chan *runable, this.taskQueueSize)
		go this.startEngine(this.runablechs[i])
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
	this.incActiveTaskNums()
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.randomExecuteTaskMode {
		this.runablechs[rand.Int()%this.engineNums] <- &runable{f: f, ps: p}
	} else {
		if this.engineIdx >= this.engineNums {
			this.engineIdx = 0
		}
		this.runablechs[this.engineIdx] <- &runable{f: f, ps: p}
		this.engineIdx++
	}
}

func (this *TaskPoolExecutor) startEngine(runablech chan *runable) {
	for !(!this.isRunning() && len(runablech) == 0) {
		//log.Infof("len(runablech):%v, this.isRunning():%v", len(runablech), this.isRunning())
		r, ok := <-runablech
		if ok {
			this.executeTask(r)
			this.decActiveTaskNums()
		} else {
			log.Warn("runable chan is closed!")
			this.decActiveTaskNums()
			return
		}
	}
	close(runablech)
	//log.Info("exit engine")
}

func (this *TaskPoolExecutor) executeTask(runable *runable) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("executeTask Runtime error caught: %v", r)
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
