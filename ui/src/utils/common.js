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

export function getCurrentConnectParams() {
  let urlParams = new URLSearchParams(window.location.search.slice(1))
  let result = {}
  urlParams.forEach(function(value, key, parent) {
    result[key] = value
  })
  return result
}