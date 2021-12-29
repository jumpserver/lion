import { del, post } from '../plugins/request'

export function createSession(url, data) {
  return post(url + '/session', data, {})
}

export function deleteSession(url, sid) {
  return del(url + `/sessions/${sid}/`, {})
}
