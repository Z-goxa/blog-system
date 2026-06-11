<template>
  <div class="p-6">
    <h2 class="text-2xl font-bold mb-6">💬 评论管理</h2>
    
    <div v-if="loading" class="text-center py-20 text-gray-400">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent text-blue-600 rounded-full mb-4" role="status"></div>
      <p>正在加载评论...</p>
    </div>

    <div v-else-if="comments.length === 0" class="text-center py-20 bg-gray-800 rounded-xl border border-gray-700">
      <p class="text-gray-400">暂无评论</p>
    </div>

    <div v-else class="bg-gray-800 rounded-xl shadow-lg overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-700">
          <tr>
            <th class="px-6 py-3 text-left text-sm">作者</th>
            <th class="px-6 py-3 text-left text-sm">内容</th>
            <th class="px-6 py-3 text-left text-sm">文章</th>
            <th class="px-6 py-3 text-left text-sm">状态</th>
            <th class="px-6 py-3 text-left text-sm">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="comment in comments" :key="comment.id" class="border-t border-gray-700 hover:bg-gray-700/50 transition">
            <td class="px-6 py-4">
              <div class="font-medium">{{ comment.author_name }}</div>
              <div class="text-xs text-gray-500">{{ comment.author_email }}</div>
            </td>
            <td class="px-6 py-4 max-w-xs truncate" :title="comment.content">{{ comment.content }}</td>
            <td class="px-6 py-4 text-sm text-gray-400">{{ comment.post_title }}</td>
            <td class="px-6 py-4">
              <span :class="statusClass(comment.status)" class="px-2 py-1 rounded text-xs">
                {{ statusText(comment.status) }}
              </span>
            </td>
            <td class="px-6 py-4">
              <button 
                v-if="comment.status === 'pending'"
                @click="handleUpdateStatus(comment.id, 'approved')" 
                class="text-green-400 hover:text-green-300 mr-3 text-sm"
              >
                通过
              </button>
              <button @click="handleDelete(comment.id)" class="text-red-400 hover:text-red-300 text-sm">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { blogAPI } from '../services/api.js'

const comments = ref([])
const loading = ref(true)

const loadComments = async () => {
  loading.value = true
  try {
    comments.value = await blogAPI.getComments()
  } catch (error) {
    console.error('加载评论失败:', error)
    comments.value = []
  } finally {
    loading.value = false
  }
}

const statusText = (status) => {
  const map = {
    'approved': '已通过',
    'pending': '待审核',
    'spam': '垃圾评论',
    'trash': '已删除'
  }
  return map[status] || status
}

const statusClass = (status) => {
  const map = {
    'approved': 'bg-green-600/20 text-green-400',
    'pending': 'bg-yellow-600/20 text-yellow-400',
    'spam': 'bg-red-600/20 text-red-400'
  }
  return map[status] || 'bg-gray-600/20 text-gray-400'
}

const handleUpdateStatus = async (id, status) => {
  try {
    await blogAPI.updateCommentStatus(id, status)
    const comment = comments.value.find(c => c.id === id)
    if (comment) comment.status = status
  } catch (error) {
    console.error('更新状态失败:', error)
    alert('操作失败: ' + error.message)
  }
}

const handleDelete = async (id) => {
  if (!confirm('确定要删除这条评论吗？')) return
  
  try {
    await blogAPI.deleteComment(id)
    comments.value = comments.value.filter(c => c.id !== id)
  } catch (error) {
    console.error('删除失败:', error)
    alert('删除失败: ' + error.message)
  }
}

onMounted(loadComments)
</script>
