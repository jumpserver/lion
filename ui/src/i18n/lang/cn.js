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
  JMSErrNoSession: '未能找到会话',
  JMSErrAuthUser: '用户未认证',
  JMSErrBadParams: '请求参数错误',
  JMSErrIdleTimeOut: '超过最大空闲时间，断开连接',
  JMSErrPermissionExpired: '授权已过期，断开连接',
  JMSErrTerminatedByAdmin: '管理员终断会话',
  JMSErrAPIFailed: 'Core API 发生错误',
  JMSErrGatewayFailed: '网关连接失败',
  JMSErrGuacamoleServer: '无法连接 Guacamole 服务器',
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
  RequireParams: '必填参数'
}

export default {
  ...el,
  ...message
}
