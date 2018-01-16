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
	state      state
	wg         sync.WaitGroup
	lock       sync.Mutex
	signalChan chan string
}

func HandleSigint() *ExecutionContext {
	executionContext := &ExecutionContext{
		state:      Running,
		signalChan: make(chan string),
	}

	go func(executionContext *ExecutionContext) {
		sigChan := make(chan os.Signal, 10)
		signal.Notify(sigChan, os.Interrupt)

		interruptCount := 0
		for ; ; {
			<-sigChan
			interruptCount++
			log.Printf("SIGINT received (%d)\n", interruptCount)

			switch interruptCount {
			case 1:
				log.Println("Stopping stream - send INT again to send INT to running docker process")
				executionContext.InitiateShutdown()
			case 2:
				log.Println("Sending INT - send INT again to send TERM running docker process")
				executionContext.signalChan <- "INT"
			case 3:
				log.Println("Sending TERM - send INT again to send KILL running docker process")
				executionContext.signalChan <- "TERM"
			case 4:
				log.Println("Sending KILL")
				executionContext.signalChan <- "KILL"
			}
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

func (executionContext *ExecutionContext) SignalChan() chan string {
	return executionContext.signalChan
}
