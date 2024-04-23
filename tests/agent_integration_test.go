package tests

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/agent"
	pb "github.com/KalashnikovProjects/ZadachaGoYaLyceum/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

type gRPCExecuteOperationTestCase struct {
	TestName string
	Request  *pb.OperationRequest
	Timeout  time.Duration
	Response *pb.OperationResponse
}

func TestAgentGRPC(t *testing.T) {
	testCases := []gRPCExecuteOperationTestCase{
		{TestName: "multiplication",
			Request: &pb.OperationRequest{
				Znak:  "*",
				Left:  0.2,
				Right: 3,
				Times: &pb.OperationTimes{Plus: 1, Minus: 1, Division: 1, Multiplication: 1}},
			Timeout:  2 * time.Second,
			Response: &pb.OperationResponse{Status: "ok", Result: 0.2 * 3},
		},
		{TestName: "zero division",
			Request: &pb.OperationRequest{
				Znak:  "/",
				Left:  4,
				Right: 0,
				Times: &pb.OperationTimes{Plus: 1, Minus: 1, Division: 1, Multiplication: 1}},
			Timeout:  2 * time.Second,
			Response: &pb.OperationResponse{Status: "error", Result: 0},
		},
		{TestName: "minus division",
			Request: &pb.OperationRequest{
				Znak:  "/",
				Left:  4.8,
				Right: -2.4,
				Times: &pb.OperationTimes{Plus: 1, Minus: 1, Division: 1, Multiplication: 1}},
			Timeout:  2 * time.Second,
			Response: &pb.OperationResponse{Status: "ok", Result: -2},
		},
		{TestName: "test time",
			Request: &pb.OperationRequest{
				Znak:  "+",
				Left:  5.4,
				Right: 4.1,
				Times: &pb.OperationTimes{Plus: 3, Minus: 5, Division: 6, Multiplication: 5}},
			Timeout:  4 * time.Second,
			Response: &pb.OperationResponse{Status: "ok", Result: 5.4 + 4.1},
		},
		{TestName: "minus",
			Request: &pb.OperationRequest{
				Znak:  "-",
				Left:  -1,
				Right: 2,
				Times: &pb.OperationTimes{Plus: 5, Minus: 3, Division: 6, Multiplication: 5}},
			Timeout:  4 * time.Second,
			Response: &pb.OperationResponse{Status: "ok", Result: -3},
		},
	}
	ctx := context.Background()
	pgContainer, err := RunPostgresContainer(ctx)
	if err != nil {
		t.Errorf("error running postgres container: %v", err)
		return
	}
	t.Cleanup(func() {
		if pgContainer.Terminate(context.Background()) != nil {
			t.Logf("error terminate postgres container: %v", err)
			return
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Errorf("error creating connection string: %v", err)
		return
	}

	t.Setenv("POSTGRES_STRING", connStr)
	t.Setenv("HMAC", "GGGGGGGGGG231241GEAW")
	t.Setenv("AGENT_COUNT", "5")

	t.Log("Запускается агент...")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go agent.ManagerAgent(ctx)

	time.Sleep(10 * time.Second)
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Errorf("error gRPC connecting: %v", err)
		return
	}
	t.Log("Запуск тестов...")
	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gRPCClient := pb.NewAgentsServiceClient(conn)
			ctx, cancel := context.WithTimeout(ctx, testCase.Timeout)
			defer cancel()
			operationResponse, err := gRPCClient.ExecuteOperation(ctx, testCase.Request)
			if err != nil {
				t.Errorf("Error gRPC ExecuteOpeation: %v", err)
				return
			}
			if operationResponse.Status != testCase.Response.Status || operationResponse.Result != testCase.Response.Result {
				t.Errorf("Operation response is not equal to expected response got: %v, want: %v", operationResponse, testCase.Response)
				return
			}
		})
	}

}
