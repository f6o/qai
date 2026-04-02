package service

import (
	"context"

	"github.com/f6o/qai/internal/gen/qaipb"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/proto/convert"
	"google.golang.org/grpc"
)

func NewRemoteServices(conn *grpc.ClientConn) (TaskService, LogService) {
	client := qaipb.NewQaiServiceClient(conn)
	return &RemoteTaskService{client: client}, &RemoteLogService{client: client}
}

type RemoteTaskService struct {
	client qaipb.QaiServiceClient
}

func (r *RemoteTaskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	resp, err := r.client.ListTasks(ctx, &qaipb.ListTasksRequest{})
	if err != nil {
		return nil, err
	}
	tasks := make([]model.Task, 0, len(resp.Tasks))
	for _, pt := range resp.Tasks {
		tasks = append(tasks, convert.ProtoToTask(pt))
	}
	return tasks, nil
}

func (r *RemoteTaskService) AddTask(ctx context.Context, task model.Task) (model.Task, error) {
	req := &qaipb.AddTaskRequest{
		Title:    task.Title,
		Status:   convert.StatusToProto(task.Status),
		Priority: int32(task.Priority),
	}
	if task.ParentID != nil {
		v := int32(*task.ParentID)
		req.ParentId = &v
	}
	resp, err := r.client.AddTask(ctx, req)
	if err != nil {
		return model.Task{}, err
	}
	return convert.ProtoToTask(resp.Task), nil
}

func (r *RemoteTaskService) UpdateTask(ctx context.Context, task model.Task) (model.Task, error) {
	resp, err := r.client.UpdateTask(ctx, &qaipb.UpdateTaskRequest{
		Task: convert.TaskToProto(task),
	})
	if err != nil {
		return model.Task{}, err
	}
	return convert.ProtoToTask(resp.Task), nil
}

func (r *RemoteTaskService) GetTask(ctx context.Context, id int) (*model.Task, error) {
	resp, err := r.client.GetTask(ctx, &qaipb.GetTaskRequest{Id: int32(id)})
	if err != nil {
		return nil, err
	}
	t := convert.ProtoToTask(resp.Task)
	return &t, nil
}

type RemoteLogService struct {
	client qaipb.QaiServiceClient
}

func (r *RemoteLogService) AppendLog(ctx context.Context, log model.Log) (model.Log, error) {
	resp, err := r.client.AppendLog(ctx, &qaipb.AppendLogRequest{
		Log: convert.LogToProto(log),
	})
	if err != nil {
		return model.Log{}, err
	}
	return convert.ProtoToLog(resp.Log), nil
}

func (r *RemoteLogService) ListLogs(ctx context.Context, opts LogListOptions) ([]model.Log, error) {
	req := &qaipb.ListLogsRequest{}
	if opts.EventType != nil {
		et := convert.EventTypeToProto(*opts.EventType)
		req.EventTypeFilter = &et
	}
	if opts.TodoID != nil {
		v := int32(*opts.TodoID)
		req.TodoIdFilter = &v
	}
	if opts.Year != nil {
		v := int32(*opts.Year)
		req.Year = &v
	}
	if opts.Month != nil {
		v := int32(*opts.Month)
		req.Month = &v
	}
	if opts.Day != nil {
		v := int32(*opts.Day)
		req.Day = &v
	}
	resp, err := r.client.ListLogs(ctx, req)
	if err != nil {
		return nil, err
	}
	logs := make([]model.Log, 0, len(resp.Logs))
	for _, pl := range resp.Logs {
		logs = append(logs, convert.ProtoToLog(pl))
	}
	return logs, nil
}
