package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeshop/tracetest/server/assertions"
	"github.com/kubeshop/tracetest/server/model"
)

type AssertionRequest struct {
	Test model.Test
	Run  model.Run
}

type AssertionRunner interface {
	RunAssertions(request AssertionRequest)
	WorkerPool
}

type defaultAssertionRunner struct {
	db           model.Repository
	inputChannel chan AssertionRequest
	exitChannel  chan bool
}

var _ WorkerPool = &defaultAssertionRunner{}
var _ AssertionRunner = &defaultAssertionRunner{}

func NewAssertionRunner(db model.Repository) AssertionRunner {
	return &defaultAssertionRunner{
		db:           db,
		inputChannel: make(chan AssertionRequest, 1),
	}
}

func (e *defaultAssertionRunner) Start(workers int) {
	e.exitChannel = make(chan bool, workers)

	for i := 0; i < workers; i++ {
		ctx := context.Background()
		go e.startWorker(ctx)
	}
}

func (e *defaultAssertionRunner) Stop() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			e.exitChannel <- true
			return
		}
	}
}

func (e *defaultAssertionRunner) startWorker(ctx context.Context) {
	for {
		select {
		case <-e.exitChannel:
			fmt.Println("Exiting assertion executor worker")
			return
		case assertionRequest := <-e.inputChannel:
			err := e.runAssertionsAndUpdateResult(ctx, assertionRequest)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func (e *defaultAssertionRunner) runAssertionsAndUpdateResult(ctx context.Context, request AssertionRequest) error {
	run, err := e.executeAssertions(ctx, request)
	if err != nil {
		return e.db.UpdateRun(ctx, run.Failed(err))
	}

	err = e.db.UpdateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("could not save result on database: %w", err)
	}

	return nil
}

func (e *defaultAssertionRunner) executeAssertions(ctx context.Context, req AssertionRequest) (model.Run, error) {
	run := req.Run
	if run.Trace == nil {
		return model.Run{}, fmt.Errorf("trace not available")
	}

	run = run.SuccessfullyAsserted(
		assertions.Assert(req.Test.Definition, *run.Trace),
	)

	return run, nil
}

func (e *defaultAssertionRunner) RunAssertions(request AssertionRequest) {
	e.inputChannel <- request
}
