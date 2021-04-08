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
if (!window.location.origin)
  streamOrigin = window.location.protocol + '//' + window.location.hostname + (window.location.port ? (':' + window.location.port) : '')
else
  streamOrigin = window.location.origin

export const OriginSite = streamOrigin

export const BaseAPIURL = streamOrigin + '/guacamole/api'
export const BaseURL = streamOrigin + '/guacamole'

const tokenBaseAPI = '/token'
const sessionBaseAPI = '/api'
const tokenWSURL = '/guacamole/ws/token/'
const wsURL = '/guacamole/ws/connect/'


export function getCurrentConnectParams() {
  let urlParams = new URLSearchParams(window.location.search.slice(1))
  let data = {}
  urlParams.forEach(function(value, key, parent) {
    data[key] = value
  })
  let result = {}
  result['data'] = data
  result['ws'] = wsURL
  result['api'] = sessionBaseAPI
  let token = urlParams.get('token')
  if (token !== null) {
    result['ws'] = tokenWSURL
    result['api'] = tokenBaseAPI
  }
  return result
}