package workers

import "context"

type RunningMode int

const (
	SYNC_BEST_MATCH RunningMode = iota
	SYNC_RECENTLY
	SYNC_MESSAGE
)

func (rm *RunningMode) GetLink() string {
	switch *rm {
	case SYNC_BEST_MATCH:
		return "https://www.upwork.com/nx/find-work/best-matches"
	case SYNC_RECENTLY:
		return "https://www.upwork.com/nx/find-work/most-recent"
	case SYNC_MESSAGE:
		return "https://www.upwork.com/ab/messages"
	default:
		return ""
	}
}

type Worker interface {
	PrepareTask() func(context.Context)
	GetMode() RunningMode
}
