package context

import (
	"sync"
	"os"
	"log"
	"os/signal"
)

type state int

const (
	Running           state = 1
	ShutdownInitiated state = 2
)

type ExecutionContext struct {
	state state
	wg    sync.WaitGroup
	lock  sync.Mutex
}

func HandleSigint() *ExecutionContext {
	executionContext := &ExecutionContext{
		state: Running,
	}

	go func(executionContext *ExecutionContext) {
		sigChan := make(chan os.Signal, 10)
		signal.Notify(sigChan, os.Interrupt)

		for ; ; {
			<-sigChan
			log.Print("SIGINT received")
			executionContext.InitiateShutdown()
		}
	}(executionContext)

	return executionContext
}

func (executionContext *ExecutionContext) InitiateShutdown() {
	executionContext.lock.Lock()
	defer executionContext.lock.Unlock()

	executionContext.state = ShutdownInitiated
}

func (executionContext *ExecutionContext) Wait() {
	executionContext.wg.Wait()
}

func (executionContext *ExecutionContext) WorkerStart() bool {
	executionContext.lock.Lock()
	defer executionContext.lock.Unlock()

	if executionContext.state == ShutdownInitiated {
		return false
	}

	executionContext.wg.Add(1)

	return true
}

func (executionContext *ExecutionContext) WorkerStop() {
	executionContext.wg.Done()
}
