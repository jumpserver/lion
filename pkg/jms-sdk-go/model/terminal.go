package model

type TerminalConfig struct {
	ReplayStorage       map[string]interface{} `json:"TERMINAL_REPLAY_STORAGE"`
	CommandStorage      map[string]interface{} `json:"TERMINAL_COMMAND_STORAGE"`
	SessionKeepDuration int                    `json:"TERMINAL_SESSION_KEEP_DURATION"`
	TelnetRegex         string                 `json:"TERMINAL_TELNET_REGEX"`
	MaxIdleTime         int                    `json:"SECURITY_MAX_IDLE_TIME"`
	HeartbeatDuration   int                    `json:"TERMINAL_HEARTBEAT_INTERVAL"`
}

type Terminal struct {
	Name           string `json:"name"`
	Comment        string `json:"comment"`
	ServiceAccount struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		AccessKey AccessKey `json:"access_key"`
	} `json:"service_account"`
}


type TerminalTask struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Args       string `json:"args"`
	IsFinished bool
}

const (
	TaskKillSession = "kill_session"
)
