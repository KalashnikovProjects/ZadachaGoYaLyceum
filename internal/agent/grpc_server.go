package agent

import (
	"context"
	"fmt"
	pb "github.com/KalashnikovProjects/ZadachaGoYaLyceum/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"os"
	"strconv"
)

type Server struct {
	pb.AgentsServiceServer // сервис из сгенерированного пакета
	tasks                  chan *TaskAgent
}

func NewServer(tasks chan *TaskAgent) *Server {
	return &Server{tasks: tasks}
}

type TaskAgent struct {
	task   *pb.OperationRequest
	result chan *pb.OperationResponse
}

func (s *Server) ExecuteOperation(
	ctx context.Context,
	in *pb.OperationRequest,
) (*pb.OperationResponse, error) {
	resp := make(chan *pb.OperationResponse)
	s.tasks <- &TaskAgent{task: in, result: resp}
	res := <-resp
	close(resp)
	return res, nil
}

func ManagerAgent(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	count, _ := strconv.Atoi(os.Getenv("AGENT_COUNT"))
	tasks := make(chan *TaskAgent) // id заданий на выполнение
	for i := 0; i < count; i++ {
		go Agent(ctx, tasks)
	}

	host := "[::]"
	port := "9090"

	addr := fmt.Sprintf("%s:%s", host, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Запущен менеджер агентов на порту", port)
	grpcServer := grpc.NewServer()
	agentsServiceServer := NewServer(tasks)

	pb.RegisterAgentsServiceServer(grpcServer, agentsServiceServer)
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return grpcServer.Serve(lis)
	})
	<-gCtx.Done()
	grpcServer.Stop()
}
