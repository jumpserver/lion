import el from 'element-ui/lib/locale/lang/zh-TW'

const message = {
  Shortcuts: '快捷鍵',
  Clipboard: '剪貼板',
  Files: '文件管理',
  WebSocketError: 'WebSocket 連接失敗，請檢查網路',
  UploadFile: '上傳文件',
  ClearDone: '清理已完成',
  Connecting: '連線中',
  ErrTitle: '連接異常',
  JMSErrPermission: '無連接權限',
  JMSErrNoSession: '未能找到會話',
  JMSErrAuthUser: '用戶未認證',
  JMSErrBadParams: '請求參數錯誤',
  JMSErrIdleTimeOut: '超過最大空閒時間{PLACEHOLDER}分鐘，斷開連接',
  JMSErrPermissionExpired: '授權已過期，斷開連接',
  JMSErrTerminatedByAdmin: '管理員終斷會話',
  JMSErrAPIFailed: 'Core API 發生錯誤',
  JMSErrGatewayFailed: '網關連接失敗',
  JMSErrGuacamoleServer: '無法連接 Guacamole 伺服器',
  JMSErrDisconnected: '會話連接已斷開',
  JMSErrMaxSession: '超過最大會話時間{PLACEHOLDER}小時，斷開連接',
  GuaErrUpstreamNotFound: '無法連接到遠程桌面伺服器（網路不可達 | 安全策略錯誤）',
  GuaErrSessionConflict: '因與另一個連接衝突，遠程桌面伺服器關閉了本連接。請稍後重試。',
  GuaErrClientUnauthorized: '使用者名稱和密碼認證錯誤，登錄失敗',
  GuaErrUnSupport: '該操作請求被禁止',
  GuaErrUpStreamTimeout: '遠程桌面伺服器無響應',
  OK: '確定',
  Submit: '提交',
  Cancel: '取消',
  Skip: '跳過',
  Username: '使用者名稱',
  Password: '密碼',
  RequireParams: '必填參數',
  Settings: '設置',
  UploadSuccess: '上傳成功',
  Display: '顯示',
  AutoFit: '自動適應',
  PauseSession: '暫停會話',
  ResumeSession: '恢復會話',

  GuacamoleErrDisconnected: '遠程連接斷開',
  GuacamoleErrCredentialsExpired: '遠程連接的憑證過期',
  GuacamoleErrSecurityNegotiationFailed: '遠程連接的安全協商失敗',
  GuacamoleErrAccessDenied: '遠程連接的訪問被拒絕',
  GuacamoleErrAuthenticationFailure: '遠程連接的認證失敗',
  GuacamoleErrSSLTLSConnectionFailed: '遠程連接的 SSL/TLS 連接失敗',
  GuacamoleErrDNSLookupFailed: '遠程連接的 DNS 查詢失敗',
  GuacamoleErrServerRefusedConnectionBySecurityType: '遠程連接的伺服器拒絕連接，可能安全類型不匹配',
  GuacamoleErrConnectionFailed: '遠程連接失敗',
  GuacamoleErrUpstreamError: '遠程連接的伺服器發生錯誤',
  GuacamoleErrForciblyDisconnected: '遠程連接被強制斷開',
  GuacamoleErrLoggedOff: '遠程連接的用戶已註銷',
  GuacamoleErrIdleSessionTimeLimitExceeded: '遠程連接的空閒時間超過限制',
  GuacamoleErrActiveSessionTimeLimitExceeded: '遠程連接的活動時間超過限制',
  GuacamoleErrDisconnectedByOtherConnection: '遠程連接被其他連接斷開',
  GuacamoleErrServerRefusedConnection: '遠程連接的伺服器拒絕連接',
  GuacamoleErrInsufficientPrivileges: '遠程連接的用戶權限不足',
  GuacamoleErrManuallyDisconnected: '遠程連接被手動斷開',
  GuacamoleErrManuallyLoggedOff: '遠程連接的用戶被手動註銷',
  GuacamoleErrUnsupportedCredentialTypeRequested: '遠程連接的憑證類型不支持'
}

export default {
  ...el,
  ...message
}