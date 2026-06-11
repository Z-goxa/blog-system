<template>
  <div class="min-h-screen bg-gray-900 text-gray-100 flex flex-col">
    <!-- 顶部提示 (HTTP API 模式) -->
    <div class="bg-green-700 text-white px-4 py-1 text-center text-xs font-bold shadow-md">
      ✅ 已连接数据库，数据实时操作
    </div>

    <div class="flex-1 flex overflow-hidden">
      <!-- 侧边栏 -->
      <aside class="w-64 bg-gray-800 border-r border-gray-700 flex flex-col">
        <!-- Logo -->
        <div class="p-6 border-b border-gray-700">
          <h1 class="text-xl font-bold text-blue-400">📝 博客管理</h1>
        </div>
        
        <!-- 导航菜单 -->
        <nav class="flex-1 p-4">
          <ul class="space-y-2">
            <li>
              <router-link to="/admin" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">🏠</span>
                <span>仪表盘</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/posts" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">📄</span>
                <span>文章管理</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/editor" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">✍️</span>
                <span>写文章</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/categories" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">📁</span>
                <span>分类管理</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/tags" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">🏷️</span>
                <span>标签管理</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/comments" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">💬</span>
                <span>评论管理</span>
              </router-link>
            </li>
            <li>
              <router-link to="/admin/users" class="flex items-center gap-3 px-4 py-3 rounded-lg hover:bg-gray-700 transition" active-class="bg-gray-700 text-blue-400">
                <span class="text-xl">👥</span>
                <span>用户管理</span>
              </router-link>
            </li>
          </ul>
        </nav>
        
        <!-- 用户信息 -->
        <div class="p-4 border-t border-gray-700">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center font-bold">
              {{ user?.username?.charAt(0).toUpperCase() || 'U' }}
            </div>
            <div class="flex-1">
              <p class="text-sm font-semibold">{{ user?.display_name || user?.username || '用户' }}</p>
              <p class="text-xs text-gray-400">{{ user?.role || '角色' }}</p>
            </div>
            <button @click="handleLogout" class="text-gray-400 hover:text-red-400 transition" title="退出登录">
              🚪
            </button>
          </div>
        </div>
      </aside>
      
      <!-- 主内容区 -->
      <main class="flex-1 overflow-auto bg-gray-900">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { user, logout } from '../services/auth.js'

const router = useRouter()

const handleLogout = () => {
  logout()
  router.push('/login')
}
</script>
