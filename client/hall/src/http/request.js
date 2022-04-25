import axios from 'axios'
import { VueAxios } from './axios'

const url = window.location.protocol + "//" +window.location.hostname
console.log(url);

// create axios
const request = axios.create({
  baseURL: url
})

// request interceptor
request.interceptors.request.use(config => {
  config.headers['Authorization'] = localStorage.getItem("token");
  return config
}, async function (error){
  return Promise.reject(error)
})

// response interceptor
request.interceptors.response.use((response) => {
  return response.data
}, async function (error) {
  return Promise.reject(error)
})


const installer = {
  vm: {},
  install (Vue) {
    Vue.use(VueAxios, request)
  }
}

export default request

export {
  installer as VueAxios,
  request as axios
}
