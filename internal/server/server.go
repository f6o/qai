package server

import (
	"context"

	"github.com/f6o/qai/internal/gen/qaipb"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/proto/convert"
	"github.com/f6o/qai/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	qaipb.UnimplementedQaiServiceServer
	tasks service.TaskService
	logs  service.LogService
}

func NewServer(tasks service.TaskService, logs service.LogService) *Server {
	return &Server{tasks: tasks, logs: logs}
}

func (s *Server) ListTasks(ctx context.Context, _ *qaipb.ListTasksRequest) (*qaipb.ListTasksResponse, error) {
	tasks, err := s.tasks.ListTasks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}
	resp := &qaipb.ListTasksResponse{}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, convert.TaskToProto(t))
	}
	return resp, nil
}

func (s *Server) AddTask(ctx context.Context, req *qaipb.AddTaskRequest) (*qaipb.AddTaskResponse, error) {
	task := model.Task{
		Title:    req.Title,
		Status:   convert.ProtoToStatus(req.Status),
		Priority: int(req.Priority),
	}
	if req.ParentId != nil {
		v := int(*req.ParentId)
		task.ParentID = &v
	}
	created, err := s.tasks.AddTask(ctx, task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add task: %v", err)
	}
	return &qaipb.AddTaskResponse{Task: convert.TaskToProto(created)}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *qaipb.UpdateTaskRequest) (*qaipb.UpdateTaskResponse, error) {
	if req.Task == nil {
		return nil, status.Error(codes.InvalidArgument, "task is required")
	}
	task := convert.ProtoToTask(req.Task)
	updated, err := s.tasks.UpdateTask(ctx, task)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}
	return &qaipb.UpdateTaskResponse{Task: convert.TaskToProto(updated)}, nil
}

func (s *Server) GetTask(ctx context.Context, req *qaipb.GetTaskRequest) (*qaipb.GetTaskResponse, error) {
	task, err := s.tasks.GetTask(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}
	if task == nil {
		return nil, status.Errorf(codes.NotFound, "task %d not found", req.Id)
	}
	return &qaipb.GetTaskResponse{Task: convert.TaskToProto(*task)}, nil
}

func (s *Server) AppendLog(ctx context.Context, req *qaipb.AppendLogRequest) (*qaipb.AppendLogResponse, error) {
	if req.Log == nil {
		return nil, status.Error(codes.InvalidArgument, "log is required")
	}
	log := convert.ProtoToLog(req.Log)
	created, err := s.logs.AppendLog(ctx, log)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to append log: %v", err)
	}
	return &qaipb.AppendLogResponse{Log: convert.LogToProto(created)}, nil
}

func (s *Server) ListLogs(ctx context.Context, req *qaipb.ListLogsRequest) (*qaipb.ListLogsResponse, error) {
	opts := service.LogListOptions{}
	if req.EventTypeFilter != nil {
		et := convert.ProtoToEventType(*req.EventTypeFilter)
		opts.EventType = &et
	}
	if req.TodoIdFilter != nil {
		v := int(*req.TodoIdFilter)
		opts.TodoID = &v
	}
	if req.Year != nil {
		v := int(*req.Year)
		opts.Year = &v
	}
	if req.Month != nil {
		v := int(*req.Month)
		opts.Month = &v
	}
	if req.Day != nil {
		v := int(*req.Day)
		opts.Day = &v
	}

	logs, err := s.logs.ListLogs(ctx, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list logs: %v", err)
	}
	resp := &qaipb.ListLogsResponse{}
	for _, l := range logs {
		resp.Logs = append(resp.Logs, convert.LogToProto(l))
	}
	return resp, nil
}
