import axios from 'axios'
import {$error} from './message'
import {BaseURL} from '@/utils/common'


const instance = axios.create({
  baseURL: BaseURL, // url = base url + request url
  withCredentials: true,
})


// 请根据实际需求修改
instance.interceptors.response.use(response => {
  return response
}, error => {
  let msg
  if (error.response) {
    msg = error.response.data.message || error.response.data
  } else {
    console.log('error: ' + error) // for debug
    msg = error.message
  }
  $error(msg)
  return Promise.reject(error)
})

export const request = instance

/* 简化请求方法，统一处理返回结果，并增加loading处理，这里以{success,data,message}格式的返回值为例，具体项目根据实际需求修改 */
const promise = (request, loading = {}) => {
  return new Promise((resolve, reject) => {
    loading.status = true
    request.then(response => {
      if (response.data.success) {
        resolve(response.data)
      } else {
        reject(response.data)
      }
      loading.status = false
    }).catch(error => {
      reject(error)
      loading.status = false
    })
  })
}

export const get = (url, data, loading) => {
  return promise(request({url: url, method: 'get', params: data}), loading)
}

export const post = (url, data, loading) => {
  return promise(request({url: url, method: 'post', data}), loading)
}

export const put = (url, data, loading) => {
  return promise(request({url: url, method: 'put', data}), loading)
}

export const del = (url, loading) => {
  return promise(request({url: url, method: 'delete'}), loading)
}

export default {
  install(Vue) {
    Vue.prototype.$get = get
    Vue.prototype.$post = post
    Vue.prototype.$put = put
    Vue.prototype.$delete = del
    Vue.prototype.$request = request
  }
}
