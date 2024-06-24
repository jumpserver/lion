import { get } from '@/plugins/request'

export function getUserInfo(query) {
  const params = {
    'action': 'suggestion',
    'search': query
  }
  return get(`/api/v1/users/users/`, params)
}
