<template>
  <div class="min-h-screen bg-gray-900 flex items-center justify-center">
    <div class="bg-gray-800 p-8 rounded-xl shadow-2xl w-96">
      <h2 class="text-2xl font-bold text-gray-100 mb-6 text-center">🔐 登录</h2>
      
      <form @submit.prevent="handleLogin">
        <div class="mb-4">
          <label class="block text-sm text-gray-400 mb-2">用户名</label>
          <input
            v-model="username"
            type="text"
            placeholder="请输入用户名"
            class="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        
        <div class="mb-6">
          <label class="block text-sm text-gray-400 mb-2">密码</label>
          <input
            v-model="password"
            type="password"
            placeholder="请输入密码"
            class="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>
        
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-blue-600 hover:bg-blue-500 text-white font-semibold py-2 rounded-lg transition disabled:opacity-50"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
        
        <p v-if="error" class="text-red-400 text-sm mt-4 text-center">{{ error }}</p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { login, isAuthenticated } from '../services/auth.js'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const router = useRouter()

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const data = await login(username.value, password.value)
    if (data.user?.role === 'subscriber') {
      error.value = '访客账号没有后台访问权限'
      router.push('/blog')
      return
    }
    // 登录成功，跳转到后台管理页面
    router.push('/admin')
  } catch (err) {
    error.value = '登录失败：' + (err.message || '未知错误')
  } finally {
    loading.value = false
  }
}

// 如果已登录且具备后台权限，在组件挂载后跳转到后台
onMounted(() => {
  if (!isAuthenticated()) return
  try {
    const savedUser = JSON.parse(localStorage.getItem('user') || '{}')
    if (savedUser.role && savedUser.role !== 'subscriber') {
      router.push('/admin')
    }
  } catch {
    // ignore broken localStorage user
  }
})
</script>
