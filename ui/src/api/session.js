import {post} from '../plugins/request'

export function createSession(data) {
  return post('/session', data, {})
}