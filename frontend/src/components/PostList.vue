<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">📄 文章管理</h2>
      <button @click="$router.push('/editor')" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-semibold transition">
        ✍️ 写新文章
      </button>
    </div>
    
    <div v-if="loading" class="text-center py-20 text-gray-400">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent text-blue-600 rounded-full mb-4" role="status"></div>
      <p>正在加载文章...</p>
    </div>

    <div v-else-if="posts.length === 0" class="text-center py-20 bg-gray-800 rounded-xl border border-gray-700">
      <p class="text-gray-400">暂无文章</p>
    </div>

    <div v-else class="bg-gray-800 rounded-xl shadow-lg overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-700">
          <tr>
            <th class="px-6 py-3 text-left text-sm">标题</th>
            <th class="px-6 py-3 text-left text-sm">分类</th>
            <th class="px-6 py-3 text-left text-sm">状态</th>
            <th class="px-6 py-3 text-left text-sm">发布时间</th>
            <th class="px-6 py-3 text-left text-sm">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="post in posts" :key="post.id" class="border-t border-gray-700 hover:bg-gray-700/50 transition">
            <td class="px-6 py-4">{{ post.title }}</td>
            <td class="px-6 py-4">
              <span class="text-gray-400">{{ post.category?.name || '无' }}</span>
            </td>
            <td class="px-6 py-4">
              <span :class="statusClass(post.status)" class="px-2 py-1 rounded text-xs">
                {{ statusText(post.status) }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-400">
              {{ post.published_at ? new Date(post.published_at).toLocaleDateString() : '-' }}
            </td>
            <td class="px-6 py-4">
              <button @click="handleEdit(post)" class="text-blue-400 hover:text-blue-300 mr-3">编辑</button>
              <button @click="handleDelete(post.id)" class="text-red-400 hover:text-red-300">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { blogAPI } from '../services/api.js'

const router = useRouter()
const posts = ref([])
const loading = ref(true)

const loadPosts = async () => {
  loading.value = true
  try {
    posts.value = await blogAPI.getAllPosts()
  } catch (error) {
    console.error('加载文章失败:', error)
    alert('加载文章失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const statusText = (status) => {
  const map = {
    'published': '已发布',
    'draft': '草稿',
    'pending': '审核中'
  }
  return map[status] || status
}

const statusClass = (status) => {
  const map = {
    'published': 'bg-green-600/20 text-green-400',
    'draft': 'bg-gray-600/20 text-gray-400',
    'pending': 'bg-yellow-600/20 text-yellow-400'
  }
  return map[status] || 'bg-blue-600/20 text-blue-400'
}

const handleEdit = (post) => {
  // 跳转到编辑器并传递 ID (后续可实现加载逻辑)
  router.push({ path: '/admin/editor', query: { id: post.id } })
}

const handleDelete = async (id) => {
  if (!confirm('确定要删除这篇文章吗？')) return
  
  try {
    await blogAPI.deletePost(id)
    posts.value = posts.value.filter(p => p.id !== id)
    alert('删除成功')
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败: ' + error.message)
  }
}

onMounted(loadPosts)
</script>
