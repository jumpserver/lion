package model

type Task struct {
	Id          string `json:"id"`
	RootDir     string `json:"root_dir"`
	DateCreated string `json:"date_created"`

	SessionId     string `json:"session_id"`
	ComponentType string `json:"component_type"`
	FileType      string `json:"file_type"`
	SessionDate   string `json:"session_date"` //  格式是 "2006-01-02"
	MaxFrame      int    `json:"max_frame"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Bitrate       int    `json:"bitrate"`

	ReplayMp4Path string `json:"replay_mp4_path"`
	ReplayPath    string `json:"replay_path"`
	Status        string `json:"status"`
	Error         string `json:"error"`
}
