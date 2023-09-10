package models

type RunningMode int

const (
	UNKNOWN RunningMode = iota
	SYNC_BEST_MATCH
	SYNC_RECENTLY
	SYNC_MESSAGE
	LOGIN_AS_CREDENTICAL
	LOGIN_AS_GOOGLE
)

func (rm *RunningMode) GetName() string {
	return []string{
		"UNKNOWN", "SYNC_BEST_MATCH", "SYNC_RECENTLY", "SYNC_MESSAGE", "LOGIN_AS_CREDENTICAL", "LOGIN_AS_GOOGLE",
	}[*rm]
}

func (rm *RunningMode) GetLink() string {
	switch *rm {
	case SYNC_BEST_MATCH:
		return "https://www.upwork.com/nx/find-work/best-matches"
	case SYNC_RECENTLY:
		return "https://www.upwork.com/nx/find-work/most-recent"
	case SYNC_MESSAGE:
		return "https://www.upwork.com/ab/messages"
	case LOGIN_AS_CREDENTICAL, LOGIN_AS_GOOGLE:
		return "https://www.upwork.com/ab/account-security/login"
	default:
		return ""
	}
}

type Config struct {
	Id      string
	Mode    RunningMode
	Account UpworkAccount
}
