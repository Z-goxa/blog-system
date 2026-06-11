<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">🏷️ 标签管理</h2>
      <div class="flex gap-2">
        <input 
          v-model="newTagName" 
          @keyup.enter="handleAddTag"
          type="text" 
          placeholder="输入新标签名称..." 
          class="bg-gray-700 border border-gray-600 rounded px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500 text-sm"
        />
        <button @click="handleAddTag" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-semibold transition">
          ➕ 添加
        </button>
      </div>
    </div>
    
    <div v-if="loading" class="text-center py-10 text-gray-400">加载中...</div>
    
    <div v-else class="bg-gray-800 rounded-xl shadow-lg p-6">
      <div v-if="tags.length > 0" class="flex flex-wrap gap-3">
        <span v-for="tag in tags" :key="tag.id" class="bg-blue-600/20 text-blue-400 px-4 py-2 rounded-full text-sm flex items-center gap-2 group">
          {{ tag.name }}
          <button @click="handleDeleteTag(tag.id)" class="hover:text-red-400 transition opacity-0 group-hover:opacity-100">&times;</button>
        </span>
      </div>
      <p v-else class="text-center text-gray-500 py-4">暂无标签</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { blogAPI } from '../services/api.js'

const tags = ref([])
const loading = ref(true)
const newTagName = ref('')

const loadTags = async () => {
  loading.value = true
  try {
    tags.value = await blogAPI.getTags()
  } catch (err) {
    console.error('加载标签失败:', err)
  } finally {
    loading.value = false
  }
}

const handleAddTag = async () => {
  if (!newTagName.value.trim()) return
  
  try {
    await blogAPI.createTag({ name: newTagName.value.trim() })
    newTagName.value = ''
    await loadTags()
  } catch (err) {
    console.error('添加标签失败:', err)
    alert('添加失败: ' + err.message)
  }
}

const handleDeleteTag = async (id) => {
  if (!confirm('确定要删除这个标签吗？')) return
  try {
    await blogAPI.deleteTag(id)
    await loadTags()
  } catch (err) {
    console.error('删除标签失败:', err)
    alert('删除失败: ' + err.message)
  }
}

onMounted(loadTags)
</script>
