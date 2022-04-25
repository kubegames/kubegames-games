import request from '@/http/request'

const hallApi = {
  login: '/hall/v1/player',
  playerInfo: '/hall/v1/player',
}

export function Login (data) {
  return request({
    url: hallApi.login,
    method: 'post',
    data: data
  })
} 

export function PlayerInfo () {
  return request({
    url: hallApi.playerInfo,
    method: 'get'
  })
}
