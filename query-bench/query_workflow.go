package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"go.temporal.io/sdk/workflow"
)

func GenerateQueryResult(payloadSize int) []byte {
	return make([]byte, payloadSize)
}

func MockActivity(ctx context.Context) (string, error) {
	return strings.Repeat(string(time.Now().Nanosecond()), 10), nil
}

func PerformanceWorkflow(ctx workflow.Context, payloadSize int) error {
	// Server for pprof
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	queryResult := GenerateQueryResult(payloadSize)

	//logger := workflow.GetLogger(ctx)
	logger.Info("PerformanceWorkflow started")
	// setup query handler for query type "state"
	err := workflow.SetQueryHandler(ctx, "state", func() (string, error) {
		return string(queryResult), nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	// populate workflow state
	for i := 0; i < 250; i++ {
		err = workflow.ExecuteActivity(ctx, MockActivity).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	for i := 0; i < 50; i++ {
		// to simulate workflow been blocked on something, in reality, workflow could wait on anything like activity, signal or timer
		_ = workflow.NewTimer(ctx, time.Second*5).Get(ctx, nil)
		logger.Info("Timer fired")

		queryResult = GenerateQueryResult(payloadSize)
	}

	_ = workflow.NewTimer(ctx, time.Minute*60).Get(ctx, nil)

	logger.Info("PerformanceWorkflow completed")
	return nil
}
