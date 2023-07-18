import el from 'element-ui/lib/locale/lang/zh-CN' // 加载element的内容

const message = {
  Shortcuts: '快捷键',
  Clipboard: '剪贴板',
  Files: '文件管理',
  WebSocketError: 'WebSocket 连接失败，请检查网络',
  UploadFile: '上传文件',
  ClearDone: '清理已完成',
  Connecting: '连接中',
  ErrTitle: '连接异常',
  JMSErrPermission: '无连接权限',
  JMSErrNoSession: '未能找到会话',
  JMSErrAuthUser: '用户未认证',
  JMSErrBadParams: '请求参数错误',
  JMSErrIdleTimeOut: '超过最大空闲时间{PLACEHOLDER}分钟，断开连接',
  JMSErrPermissionExpired: '授权已过期，断开连接',
  JMSErrTerminatedByAdmin: '管理员终断会话',
  JMSErrAPIFailed: 'Core API 发生错误',
  JMSErrGatewayFailed: '网关连接失败',
  JMSErrGuacamoleServer: '无法连接 Guacamole 服务器',
  JMSErrDisconnected: '会话连接已断开',
  GuaErrUpstreamNotFound: '无法连接到远程桌面服务器（网络不可达 | 安全策略错误）',
  GuaErrSessionConflict: '因与另一个连接冲突，远程桌面服务器关闭了本连接。请稍后重试。',
  GuaErrClientUnauthorized: '用户名和密码认证错误，登录失败',
  GuaErrUnSupport: '该操作请求被禁止',
  GuaErrUpStreamTimeout: '远程桌面服务器无响应',
  OK: '确定',
  Submit: '提交',
  Cancel: '取消',
  Skip: '跳过',
  Username: '用户名',
  Password: '密码',
  RequireParams: '必填参数',
  Settings: '设置',
  UploadSuccess: '上传成功',
  Display: '显示',
  AutoFit: '自动适应',

  GuacamoleErrDisconnected: '远程连接断开',
  GuacamoleErrCredentialsExpired: '远程连接的凭证过期',
  GuacamoleErrSecurityNegotiationFailed: '远程连接的安全协商失败',
  GuacamoleErrAccessDenied: '远程连接的访问被拒绝',
  GuacamoleErrAuthenticationFailure: '远程连接的认证失败',
  GuacamoleErrSSLTLSConnectionFailed: '远程连接的 SSL/TLS 连接失败',
  GuacamoleErrDNSLookupFailed: '远程连接的 DNS 查询失败',
  GuacamoleErrServerRefusedConnectionBySecurityType: '远程连接的服务器拒绝连接，可能安全类型不匹配',
  GuacamoleErrConnectionFailed: '远程连接失败',
  GuacamoleErrUpstreamError: '远程连接的服务器发生错误',
  GuacamoleErrForciblyDisconnected: '远程连接被强制断开',
  GuacamoleErrLoggedOff: '远程连接的用户已注销',
  GuacamoleErrIdleSessionTimeLimitExceeded: '远程连接的空闲时间超过限制',
  GuacamoleErrActiveSessionTimeLimitExceeded: '远程连接的活动时间超过限制',
  GuacamoleErrDisconnectedByOtherConnection: '远程连接被其他连接断开',
  GuacamoleErrServerRefusedConnection: '远程连接的服务器拒绝连接',
  GuacamoleErrInsufficientPrivileges: '远程连接的用户权限不足',
  GuacamoleErrManuallyDisconnected: '远程连接被手动断开',
  GuacamoleErrManuallyLoggedOff: '远程连接的用户被手动注销',
  GuacamoleErrUnsupportedCredentialTypeRequested: '远程连接的凭证类型不支持'
}

export default {
  ...el,
  ...message
}
