package convert

import (
	"github.com/f6o/qai/internal/gen/qaipb"
	"github.com/f6o/qai/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StatusToProto(s model.Status) qaipb.Status {
	switch s {
	case model.StatusIdea:
		return qaipb.Status_STATUS_IDEA
	case model.StatusTodo:
		return qaipb.Status_STATUS_TODO
	case model.StatusDoing:
		return qaipb.Status_STATUS_DOING
	case model.StatusDone:
		return qaipb.Status_STATUS_DONE
	default:
		return qaipb.Status_STATUS_UNSPECIFIED
	}
}

func ProtoToStatus(s qaipb.Status) model.Status {
	switch s {
	case qaipb.Status_STATUS_IDEA:
		return model.StatusIdea
	case qaipb.Status_STATUS_TODO:
		return model.StatusTodo
	case qaipb.Status_STATUS_DOING:
		return model.StatusDoing
	case qaipb.Status_STATUS_DONE:
		return model.StatusDone
	default:
		return ""
	}
}

func EventTypeToProto(e model.EventType) qaipb.EventType {
	switch e {
	case model.EventFocusComplete:
		return qaipb.EventType_EVENT_TYPE_FOCUS_COMPLETE
	case model.EventFocusSkip:
		return qaipb.EventType_EVENT_TYPE_FOCUS_SKIP
	case model.EventFocusQuit:
		return qaipb.EventType_EVENT_TYPE_FOCUS_QUIT
	case model.EventTaskCreate:
		return qaipb.EventType_EVENT_TYPE_TASK_CREATE
	case model.EventTaskContinue:
		return qaipb.EventType_EVENT_TYPE_TASK_CONTINUE
	case model.EventStatusChange:
		return qaipb.EventType_EVENT_TYPE_STATUS_CHANGE
	default:
		return qaipb.EventType_EVENT_TYPE_UNSPECIFIED
	}
}

func ProtoToEventType(e qaipb.EventType) model.EventType {
	switch e {
	case qaipb.EventType_EVENT_TYPE_FOCUS_COMPLETE:
		return model.EventFocusComplete
	case qaipb.EventType_EVENT_TYPE_FOCUS_SKIP:
		return model.EventFocusSkip
	case qaipb.EventType_EVENT_TYPE_FOCUS_QUIT:
		return model.EventFocusQuit
	case qaipb.EventType_EVENT_TYPE_TASK_CREATE:
		return model.EventTaskCreate
	case qaipb.EventType_EVENT_TYPE_TASK_CONTINUE:
		return model.EventTaskContinue
	case qaipb.EventType_EVENT_TYPE_STATUS_CHANGE:
		return model.EventStatusChange
	default:
		return ""
	}
}

func TaskToProto(t model.Task) *qaipb.Task {
	pt := &qaipb.Task{
		Id:       int32(t.ID),
		Title:    t.Title,
		Status:   StatusToProto(t.Status),
		Priority: int32(t.Priority),
	}
	if t.ParentID != nil {
		v := int32(*t.ParentID)
		pt.ParentId = &v
	}
	if !t.StartedAt.IsZero() {
		pt.StartedAt = timestamppb.New(t.StartedAt)
	}
	if !t.CreatedAt.IsZero() {
		pt.CreatedAt = timestamppb.New(t.CreatedAt)
	}
	return pt
}

func ProtoToTask(pt *qaipb.Task) model.Task {
	t := model.Task{
		ID:       int(pt.Id),
		Title:    pt.Title,
		Status:   ProtoToStatus(pt.Status),
		Priority: int(pt.Priority),
	}
	if pt.ParentId != nil {
		v := int(*pt.ParentId)
		t.ParentID = &v
	}
	if pt.StartedAt != nil {
		t.StartedAt = pt.StartedAt.AsTime()
	}
	if pt.CreatedAt != nil {
		t.CreatedAt = pt.CreatedAt.AsTime()
	}
	return t
}

func LogToProto(l model.Log) *qaipb.Log {
	pl := &qaipb.Log{
		Id:         int32(l.ID),
		TodoId:     int32(l.TodoID),
		Content:    l.Content,
		EventType:  EventTypeToProto(l.EventType),
		FromStatus: StatusToProto(l.FromStatus),
		ToStatus:   StatusToProto(l.ToStatus),
	}
	if l.Duration != nil {
		v := int32(*l.Duration)
		pl.Duration = &v
	}
	if !l.LoggedAt.IsZero() {
		pl.LoggedAt = timestamppb.New(l.LoggedAt)
	}
	return pl
}

func ProtoToLog(pl *qaipb.Log) model.Log {
	l := model.Log{
		ID:         int(pl.Id),
		TodoID:     int(pl.TodoId),
		Content:    pl.Content,
		EventType:  ProtoToEventType(pl.EventType),
		FromStatus: ProtoToStatus(pl.FromStatus),
		ToStatus:   ProtoToStatus(pl.ToStatus),
	}
	if pl.Duration != nil {
		v := int(*pl.Duration)
		l.Duration = &v
	}
	if pl.LoggedAt != nil {
		l.LoggedAt = pl.LoggedAt.AsTime()
	}
	return l
}
