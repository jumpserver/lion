import el from 'element-ui/lib/locale/lang/zh-CN' // 加载element的内容

const message = {
  Shortcuts: '快捷键',
  Clipboard: '剪贴板',
  Files: '文件管理',
  WebSocketError: 'WebSocket 连接失败，请检查网络',
  UploadFile: '上传文件',
  ClearDone: '清理已完成',
  Connecting: '连接中'
}

export default {
  ...el,
  ...message
}
