package service

// 与Core交互的API
const (
	UserProfileURL       = "/api/v1/users/profile/"                   // 获取当前用户的基本信息
	TerminalRegisterURL  = "/api/v1/terminal/terminal-registrations/" // 注册
	TerminalConfigURL    = "/api/v1/terminal/terminals/config/"       // 获取配置
	TerminalHeartBeatURL = "/api/v1/terminal/terminals/status/"
)

// 用户登陆认证使用的API
const (
	TokenAssetURL = "/api/v1/authentication/connection-token/%s/" // Token name
)

// Session相关API
const (
	SessionListURL      = "/api/v1/terminal/sessions/"           //上传创建的资产会话session id
	SessionDetailURL    = "/api/v1/terminal/sessions/%s/"        // finish session的时候发送
	SessionReplayURL    = "/api/v1/terminal/sessions/%s/replay/" //上传录像
	SessionCommandURL   = "/api/v1/terminal/commands/"           //上传批量命令
	FinishTaskURL       = "/api/v1/terminal/tasks/%s/"
	JoinRoomValidateURL = "/api/v1/terminal/sessions/join/validate/"
	FTPLogListURL       = "/api/v1/audits/ftp-logs/" // 上传 ftp日志
	FTPLogFileURL       = "/api/v1/audits/ftp-logs/%s/upload/"
	FTPLogUpdateURL     = "/api/v1/audits/ftp-logs/%s/"

	SessionLifecycleLogURL = "/api/v1/terminal/sessions/%s/lifecycle_log/"
)

// 各资源详情相关API
const (
	UserDetailURL    = "/api/v1/users/users/%s/"
	AssetDetailURL   = "/api/v1/assets/assets/%s/"
	AssetPlatFormURL = "/api/v1/assets/assets/%s/platform/"

	DomainDetailWithGateways = "/api/v1/assets/domains/%s/?gateway=1"
)

const (
	TicketSessionURL = "/api/v1/tickets/ticket-session-relation/"
)

const (
	SuperConnectTokenSecretURL = "/api/v1/authentication/super-connection-token/secret/"

	SuperConnectTokenAppletOptionURL        = "/api/v1/authentication/super-connection-token/applet-option/"
	SuperConnectAppletHostAccountReleaseURL = "/api/v1/authentication/super-connection-token/applet-account/release/"

	SuperConnectTokenVirtualAppOptionURL = "/api/v1/authentication/super-connection-token/virtual-app-option/"

	SuperConnectTokenCheckURL = "/api/v1/authentication/super-connection-token/%s/check/"
)

const (
	ShareCreateURL        = "/api/v1/terminal/session-sharings/"
	ShareSessionJoinURL   = "/api/v1/terminal/session-join-records/"
	ShareSessionFinishURL = "/api/v1/terminal/session-join-records/%s/finished/"
)
