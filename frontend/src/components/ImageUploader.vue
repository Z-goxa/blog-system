<template>
  <div class="relative">
    <!-- 上传按钮 -->
    <button
      @click="triggerUpload"
      type="button"
      class="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition flex items-center gap-2"
    >
      📷 上传图片
    </button>
    
    <!-- 隐藏的文件输入 -->
    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      multiple
      class="hidden"
      @change="handleFileSelect"
    />
    
    <!-- 拖拽覆盖层 -->
    <div
      v-if="isDragging"
      class="fixed inset-0 bg-blue-500/20 border-4 border-blue-500 border-dashed z-50 flex items-center justify-center"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      @drop.prevent="onDrop"
    >
      <p class="text-2xl font-bold text-blue-500">松开鼠标上传图片</p>
    </div>
    
    <!-- 上传进度 -->
    <div v-if="uploading" class="mt-2 text-sm text-gray-400">
      上传中...
    </div>
    
    <!-- 上传结果 -->
    <div v-if="uploadResult" class="mt-2 text-sm text-green-400">
      ✅ 上传成功: {{ uploadResult.name }}
    </div>
    
    <!-- 错误提示 -->
    <div v-if="error" class="mt-2 text-sm text-red-400">
      ❌ {{ error }}
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { UploadImage } from '../wailsjs/go/main/UploadService.js'
import { blogAPI } from '../services/api.js'

// 检测 Wails 环境
const isWails = () => window.go && window.go.main

const fileInput = ref(null)
const isDragging = ref(false)
const uploading = ref(false)
const uploadResult = ref(null)
const error = ref('')

// 触发文件选择
const triggerUpload = () => {
  fileInput.value.click()
}

// 处理文件选择
const handleFileSelect = async (event) => {
  const files = event.target.files
  if (files.length > 0) {
    await uploadFiles(files)
  }
}

// 拖拽事件
const onDragOver = (e) => {
  isDragging.value = true
}

const onDragLeave = (e) => {
  isDragging.value = false
}

const onDrop = async (e) => {
  isDragging.value = false
  const files = e.dataTransfer.files
  if (files.length > 0) {
    await uploadFiles(files)
  }
}

// 上传文件
const uploadFiles = async (files) => {
  uploading.value = true
  error.value = ''
  uploadResult.value = null
  
  try {
    for (const file of files) {
      if (!file.type.startsWith('image/')) {
        error.value = '只支持图片文件'
        continue
      }
      
      let result
      if (isWails()) {
        const buffer = await file.arrayBuffer()
        const bytes = Array.from(new Uint8Array(buffer))
        result = await UploadImage(file.name, bytes)
      } else {
        const res = await blogAPI.uploadImage(file)
        if (!res.ok) {
          const err = await res.json().catch(() => ({}))
          throw new Error(err.error || '上传失败')
        }
        result = await res.json()
      }

      uploadResult.value = result
      emit('upload-success', result)
    }
  } catch (err) {
    console.error('上传失败:', err)
    error.value = err.message || '上传失败'
  } finally {
    uploading.value = false
  }
}

// 定义事件
const emit = defineEmits(['upload-success'])
</script>
