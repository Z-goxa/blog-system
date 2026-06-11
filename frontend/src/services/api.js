// HTTP API 服务
import { getToken } from './auth.js'

// API 基础配置
export const API_BASE = import.meta.env.VITE_API_URL || '/api'

// 请求拦截器
const request = async (url, options = {}) => {
  const token = getToken()
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  
  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  })
  
  if (!response.ok) {
    let errorMessage
    try {
      const errorData = await response.json()
      errorMessage = errorData.error || `HTTP ${response.status} 错误`
    } catch {
      const errorText = await response.text()
      errorMessage = `HTTP ${response.status}: ${errorText}`
    }
    throw new Error(errorMessage)
  }
  
  return response.json()
}

// GET 请求
const get = (url, options = {}) => {
  return request(url, { ...options, method: 'GET' })
}

// POST 请求
const post = (url, data, options = {}) => {
  return request(url, {
    ...options,
    method: 'POST',
    body: JSON.stringify(data),
  })
}

// PUT 请求
const put = (url, data, options = {}) => {
  return request(url, {
    ...options,
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

// DELETE 请求
const del = (url, options = {}) => {
  return request(url, { ...options, method: 'DELETE' })
}

// 博客服务 API
const blogAPI = {
  // 获取统计数据
  getStats: () => get('/stats'),
  
  // 获取所有文章
  getAllPosts: () => get('/posts'),
  
  // 获取单个文章
  getPost: (id) => get(`/posts/${id}`),
  
  // 创建文章
  createPost: (data) => post('/posts', data),
  
  // 更新文章
  updatePost: (id, data) => put(`/posts/${id}`, data),
  
  // 删除文章
  deletePost: (id) => del(`/posts/${id}`),
  
  // 获取所有分类
  getCategories: () => get('/categories'),
  
  // 创建分类
  createCategory: (data) => post('/categories', data),
  
  // 更新分类
  updateCategory: (id, data) => put(`/categories/${id}`, data),
  
  // 删除分类
  deleteCategory: (id) => del(`/categories/${id}`),
  
  // 获取所有标签
  getTags: () => get('/tags'),
  
  // 创建标签
  createTag: (data) => post('/tags', data),
  
  // 更新标签
  updateTag: (id, data) => put(`/tags/${id}`, data),
  
  // 删除标签
  deleteTag: (id) => del(`/tags/${id}`),
  
  // 获取用户列表
  getUsers: () => get('/users'),

  // 强制修改用户密码
  forceUpdateUserPassword: (id, password) => put(`/users/${id}/password`, { password }),

  // 修改用户角色
  updateUserRole: (id, role) => put(`/users/${id}/role`, { role }),

  // 删除用户
  deleteUser: (id) => del(`/users/${id}`),
  
  // 获取所有评论
  getComments: () => get('/comments'),
  
  // 回复评论
  replyComment: (id, data) => post(`/comments/${id}/reply`, data),
  
  // 删除评论
  deleteComment: (id) => del(`/comments/${id}`),
  
  // 更新评论状态
  updateCommentStatus: (id, status) => put(`/comments/${id}/status`, { status }),
  
  // 上传图片
  uploadImage: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    
    return fetch(`${API_BASE}/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getToken()}`,
      },
      body: formData,
    })
  },
  
  // 搜索文章
  searchPosts: (query) => get(`/posts/search?q=${encodeURIComponent(query)}`),
}

export {
  get,
  post,
  put,
  del,
  blogAPI,
}
