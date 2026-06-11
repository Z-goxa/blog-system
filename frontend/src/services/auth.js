// 认证服务
import { ref } from 'vue'
import { API_BASE } from './api.js'

const token = ref(localStorage.getItem('token') || '')
const user = ref(null)

// 安全地解析用户信息
try {
  const savedUser = localStorage.getItem('user')
  if (savedUser) {
    user.value = JSON.parse(savedUser)
  }
} catch (e) {
  console.error('解析本地用户信息失败:', e)
  localStorage.removeItem('user')
}

// 登录
const login = async (username, password) => {
  try {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username,
        password,
      }),
    })
    
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || '登录失败')
    }
    
    const data = await response.json()
    token.value = data.token
    user.value = data.user
    
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    
    return data
  } catch (error) {
    console.error('登录失败:', error)
    throw error
  }
}

// 登出
const logout = () => {
  token.value = ''
  user.value = null
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

// 检查是否已登录
const isAuthenticated = () => {
  return !!token.value
}

// 获取 Token
const getToken = () => {
  return token.value
}

export {
  token,
  user,
  login,
  logout,
  isAuthenticated,
  getToken,
}
