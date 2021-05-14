import el from 'element-ui/lib/locale/lang/zh-CN' // 加载element的内容

const message = {
  Shortcuts: '快捷键',
  Clipboard: '剪贴板',
  Files: '文件管理',
  WebSocketError: 'WebSocket 连接失败，请检查网络'
}

export default {
  ...el,
  ...message
}
