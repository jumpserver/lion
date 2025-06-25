export function sanitizeFilename(filename: string): string {
  return filename.replace(/[\\\/]+/g, '_')
}

export const FileType = {
  NORMAL: 'NORMAL',
  DIRECTORY: 'DIRECTORY'
}

export function isDirectory(guacFile: { type: string }): boolean {
  return guacFile.type === FileType.DIRECTORY
}

let streamOrigin
// Work-around for IE missing window.location.origin
if (!window.location.origin) {
  streamOrigin = window.location.protocol + '//' + window.location.hostname + (window.location.port ? (':' + window.location.port) : '')
} else {
  streamOrigin = window.location.origin
}
const scheme = document.location.protocol === 'https:' ? 'wss' : 'ws'
const port = document.location.port ? ':' + document.location.port : ''
const BASE_WS_URL = scheme + '://' + document.location.hostname + port
const BASE_URL = document.location.protocol + '//' + document.location.hostname + port
export { BASE_WS_URL, BASE_URL }

export const OriginSite = streamOrigin

export const BaseAPIURL = streamOrigin + '/lion/api'

const sessionBaseAPI = '/api'
const wsURL = '/lion/ws/connect/'
const monitorWsURL = '/lion/ws/monitor/'

export function getCurrentConnectParams() {
  const urlParams = getURLParams()
  const data: any = {}
  urlParams.forEach(function (value, key, parent) {
    data[key] = value
  })
  const result: any = {}
  result['data'] = data
  result['ws'] = wsURL
  result['api'] = sessionBaseAPI
  return result
}

export function getMonitorConnectParams() {
  const urlParams = getURLParams()
  const data: any = {}
  urlParams.forEach(function (value, key, parent) {
    data[key] = value
  })
  const result: any = {}
  result['data'] = data
  result['ws'] = monitorWsURL
  return result
}

export function getURLParams() {
  return new URLSearchParams(window.location.search.slice(1))
}

export function localStorageGet(key: string): string | object | null {
  let data = localStorage.getItem(key)
  if (!data) {
    return data
  }
  try {
    data = JSON.parse(data)
    return data
  } catch (e) {
    //
  }
  return data
}

export function getCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp(name + '=([^;]+)'))
  return match ? match[1] : undefined
}


export function CopyTextToClipboard(text: string) {
  const transfer = document.createElement('textarea')
  document.body.appendChild(transfer)
  transfer.value = text
  transfer.focus()
  transfer.select()
  document.execCommand('copy')
  document.body.removeChild(transfer)
}
