<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">📊 仪表盘</h2>
      <button 
        @click="loadDashboardData" 
        class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition flex items-center gap-2"
        :disabled="loading"
      >
        <span>{{ loading ? '🔄 刷新中...' : '🔄 刷新数据' }}</span>
      </button>
    </div>
    
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <!-- 统计卡片 -->
      <div class="bg-gray-800 p-6 rounded-xl shadow-lg">
        <div class="flex items-center justify-between mb-4">
          <span class="text-gray-400">文章总数</span>
          <span class="text-3xl">📄</span>
        </div>
        <p class="text-3xl font-bold">{{ stats.post_count }}</p>
      </div>
      
      <div class="bg-gray-800 p-6 rounded-xl shadow-lg">
        <div class="flex items-center justify-between mb-4">
          <span class="text-gray-400">分类数</span>
          <span class="text-3xl">📁</span>
        </div>
        <p class="text-3xl font-bold">{{ stats.category_count }}</p>
      </div>
      
      <div class="bg-gray-800 p-6 rounded-xl shadow-lg">
        <div class="flex items-center justify-between mb-4">
          <span class="text-gray-400">标签数</span>
          <span class="text-3xl">🏷️</span>
        </div>
        <p class="text-3xl font-bold">{{ stats.tag_count }}</p>
      </div>
      
      <div class="bg-gray-800 p-6 rounded-xl shadow-lg">
        <div class="flex items-center justify-between mb-4">
          <span class="text-gray-400">评论数</span>
          <span class="text-3xl">💬</span>
        </div>
        <p class="text-3xl font-bold">{{ stats.comment_count }}</p>
      </div>
    </div>
    
    <!-- 最近文章 -->
    <div class="bg-gray-800 rounded-xl shadow-lg p-6">
      <h3 class="text-xl font-semibold mb-4">📝 最近文章</h3>
      <div v-if="loading" class="text-gray-400 py-4">加载中...</div>
      <div v-else-if="recentPosts.length === 0" class="text-gray-400 py-4">暂无文章</div>
      <ul v-else class="space-y-3">
        <li v-for="post in recentPosts" :key="post.id" @click="$router.push({path: '/admin/editor', query: {id: post.id}})" class="flex items-center justify-between p-3 hover:bg-gray-700 rounded-lg cursor-pointer transition">
          <span>{{ post.title }}</span>
          <span class="text-sm text-gray-400">{{ new Date(post.created_at).toLocaleDateString() }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { blogAPI } from '../services/api.js'

const stats = ref({
  post_count: 0,
  category_count: 0,
  tag_count: 0,
  comment_count: 0
})
const recentPosts = ref([])
const loading = ref(true)
const errorInfo = ref('')

const loadDashboardData = async () => {
  loading.value = true
  errorInfo.value = ''
  try {
    const [s, posts] = await Promise.all([
      blogAPI.getStats(),
      blogAPI.getAllPosts()
    ])
    stats.value = s
    recentPosts.value = posts.slice(0, 5)
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
    errorInfo.value = `加载失败: ${error.message || error}`
    alert('加载失败: ' + (error.message || error))
  } finally {
    loading.value = false
  }
}

onMounted(loadDashboardData)
</script>
