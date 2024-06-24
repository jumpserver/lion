import { post } from '@/plugins/request'

export function createShareURL(data) {
  return post(`/lion/api/share/`, data)
}

export function getShareSession(id, data) {
  return post(`/lion/api/share/${id}/`, data)
}
