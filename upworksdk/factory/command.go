package factory

type CommandType int

const (
	START_CMD CommandType = iota
	UPDATE_JOB_CMD
	STOP_CMD
)

func (ct CommandType) Name() string {
	return []string{"START_CMD", "UPDATE_JOB_CMD", "STOP_CMD"}[ct]
}

type Command struct {
	Type CommandType
	Data interface{}
}
