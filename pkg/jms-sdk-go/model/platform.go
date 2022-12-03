package model

type Platform struct {
	BaseOs string `json:"base"`
	ID     int    `json:"id"`
	Name   string `json:"name"`

	Protocols     []PlatformProtocol `json:"protocols"`
	Category      LabelValue         `json:"category"`
	Charset       LabelValue         `json:"charset"`
	Type          LabelValue         `json:"type"`
	SuEnabled     bool               `json:"su_enabled"`
	SuMethod      string             `json:"su_method"`
	DomainEnabled bool               `json:"domain_enabled"`
	Comment       string             `json:"comment"`
}

type PlatformProtocol struct {
	Protocol
	ProtocolSetting
}
type ProtocolSetting struct {
	Security         string `json:"security"`
	SftpEnabled      bool   `json:"sftp_enabled"`
	SftpHome         string `json:"sftp_home"`
	AutoFill         bool   `json:"auto_fill"`
	UsernameSelector string `json:"username_selector"`
	PasswordSelector string `json:"password_selector"`
	SubmitSelector   string `json:"submit_selector"`
}
