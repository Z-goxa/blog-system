<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">📁 分类管理</h2>
      <button 
        @click="handleAdd"
        class="px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-semibold transition"
      >
        ➕ 添加分类
      </button>
    </div>
    
    <div v-if="loading" class="text-center py-10 text-gray-400">加载中...</div>
    
    <div v-else class="bg-gray-800 rounded-xl shadow-lg p-6">
      <ul v-if="categories.length > 0" class="space-y-3">
        <li v-for="cat in categories" :key="cat.id" class="flex items-center justify-between p-3 hover:bg-gray-700 rounded-lg transition">
          <div class="flex items-center gap-3">
            <span class="text-xl">📁</span>
            <div>
              <span class="font-semibold">{{ cat.name }}</span>
              <p class="text-xs text-gray-400">{{ cat.slug }}</p>
            </div>
          </div>
          <div class="flex gap-3">
            <button @click="handleEdit(cat)" class="text-blue-400 hover:text-blue-300 text-sm">编辑</button>
            <button @click="handleDelete(cat.id)" class="text-red-400 hover:text-red-300 text-sm">删除</button>
          </div>
        </li>
      </ul>
      <p v-else class="text-center text-gray-500 py-4">暂无分类</p>
    </div>

    <!-- 添加/编辑弹窗 -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-gray-800 p-6 rounded-xl w-96 shadow-2xl">
        <h3 class="text-xl font-bold mb-4">{{ currentCategory.id ? '编辑分类' : '添加分类' }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-gray-400 mb-1">名称</label>
            <input v-model="currentCategory.name" type="text" class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">标识 (Slug)</label>
            <input v-model="currentCategory.slug" type="text" class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500" placeholder="留空则自动生成" />
          </div>
          <div>
            <label class="block text-sm text-gray-400 mb-1">描述</label>
            <textarea v-model="currentCategory.description" class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500 h-20"></textarea>
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false" class="px-4 py-2 text-gray-400 hover:text-gray-200 transition">取消</button>
          <button @click="handleSave" class="px-6 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg font-semibold transition">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { blogAPI } from '../services/api.js'

const categories = ref([])
const loading = ref(true)
const showModal = ref(false)
const currentCategory = ref({ name: '', slug: '', description: '' })

const loadCategories = async () => {
  loading.value = true
  try {
    categories.value = await blogAPI.getCategories()
  } catch (err) {
    console.error('加载分类失败:', err)
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  currentCategory.value = { name: '', slug: '', description: '' }
  showModal.value = true
}

const handleEdit = (cat) => {
  currentCategory.value = { ...cat }
  showModal.value = true
}

const handleSave = async () => {
  if (!currentCategory.value.name) {
    alert('名称不能为空')
    return
  }
  
  loading.value = true
  try {
    if (currentCategory.value.id) {
      await blogAPI.updateCategory(currentCategory.value.id, {
        name: currentCategory.value.name,
        slug: currentCategory.value.slug || '',
        description: currentCategory.value.description || ''
      })
    } else {
      await blogAPI.createCategory({
        name: currentCategory.value.name,
        slug: currentCategory.value.slug || '',
        description: currentCategory.value.description || ''
      })
    }
    alert('保存成功')
    
    showModal.value = false
    await loadCategories()
  } catch (err) {
    console.error('保存失败:', err)
    alert('保存失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

const handleDelete = async (id) => {
  if (!confirm('确定要删除这个分类吗？')) return
  try {
    await blogAPI.deleteCategory(id)
    await loadCategories()
    alert('删除成功')
  } catch (err) {
    console.error('删除失败:', err)
    alert('删除失败: ' + err.message)
  }
}

onMounted(loadCategories)
</script>
