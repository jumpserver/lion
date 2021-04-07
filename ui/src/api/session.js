import {post} from '../plugins/request'

export function createSession(url, data) {
  return post(url, data, {})
}