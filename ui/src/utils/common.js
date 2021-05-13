export function sanitizeFilename(filename) {
  return filename.replace(/[\\\/]+/g, '_')
}

export const FileType = {
  NORMAL: 'NORMAL',
  DIRECTORY: 'DIRECTORY'
}

export function isDirectory(guacFile) {
  return guacFile.type === FileType.DIRECTORY
}

let streamOrigin
// Work-around for IE missing window.location.origin
if (!window.location.origin) {
  streamOrigin = window.location.protocol + '//' + window.location.hostname + (window.location.port ? (':' + window.location.port) : '')
} else {
  streamOrigin = window.location.origin
}

export const OriginSite = streamOrigin

export const BaseAPIURL = streamOrigin + '/lion/api'
export const BaseURL = streamOrigin + '/lion'
export const apiPrefix = '/api'

const tokenBaseAPI = '/token'
const sessionBaseAPI = '/api'
const tokenWSURL = '/lion/ws/token/'
const wsURL = '/lion/ws/connect/'
const monitorWsURL = '/lion/ws/monitor/'

export function getCurrentConnectParams() {
  const urlParams = getURLParams()
  const data = {}
  urlParams.forEach(function(value, key, parent) {
    data[key] = value
  })
  const result = {}
  result['data'] = data
  result['ws'] = wsURL
  result['api'] = sessionBaseAPI
  const token = urlParams.get('token')
  if (token !== null) {
    result['ws'] = tokenWSURL
    result['api'] = tokenBaseAPI
  }
  return result
}

export function getMonitorConnectParams() {
  const urlParams = getURLParams()
  const data = {}
  urlParams.forEach(function(value, key, parent) {
    data[key] = value
  })
  const result = {}
  result['data'] = data
  result['ws'] = monitorWsURL
  return result
}

function getURLParams() {
  return new URLSearchParams(window.location.search.slice(1))
}
