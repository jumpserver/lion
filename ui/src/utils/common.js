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

/*
const tokenBaseAPI = '/token'
const tokenWSURL = '/lion/ws/token/'

 */

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
  // todo: 后续支持不登录直接连接的方式
  // const token = urlParams.get('token')
  // if (token !== null) {
  //   result['ws'] = tokenWSURL
  //   result['api'] = tokenBaseAPI
  // }
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

export function getURLParams() {
  return new URLSearchParams(window.location.search.slice(1))
}

export function localStorageGet(key) {
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

export function getCookie(name) {
  const match = document.cookie.match(new RegExp(name + '=([^;]+)'))
  return match ? match[1] : undefined
}

/**
 * Check if an element has a class
 * @param {HTMLElement} elm
 * @param {string} cls
 * @returns {boolean}
 */
export function hasClass(ele, cls) {
  return !!ele.className.match(new RegExp('(\\s|^)' + cls + '(\\s|$)'))
}

/**
 * Add class to element
 * @param {HTMLElement} elm
 * @param {string} cls
 */
export function addClass(ele, cls) {
  if (!hasClass(ele, cls)) ele.className += ' ' + cls
}

/**
 * Remove class from element
 * @param {HTMLElement} elm
 * @param {string} cls
 */
export function removeClass(ele, cls) {
  if (hasClass(ele, cls)) {
    const reg = new RegExp('(\\s|^)' + cls + '(\\s|$)')
    ele.className = ele.className.replace(reg, ' ')
  }
}

export function CopyTextToClipboard(text) {
  const transfer = document.createElement('textarea')
  document.body.appendChild(transfer)
  transfer.value = text
  transfer.focus()
  transfer.select()
  document.execCommand('copy')
  document.body.removeChild(transfer)
}
