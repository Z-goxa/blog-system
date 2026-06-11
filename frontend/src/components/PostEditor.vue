<template>
  <div class="min-h-screen bg-gray-900 text-gray-100 flex flex-col">
    <!-- 顶部工具栏 -->
    <header class="bg-gray-800 border-b border-gray-700 px-6 py-4 flex items-center justify-between shadow-lg">
      <div class="flex items-center gap-4 flex-1">
        <div class="hidden md:flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-blue-600/20 text-blue-300 border border-blue-500/30">
          {{ postId ? '编辑文章' : '发布新文章' }}
        </div>
        <!-- 标题输入 -->
        <input
          v-model="form.title"
          @input="handleTitleChange"
          type="text"
          placeholder="输入文章标题..."
          class="flex-1 bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-lg font-semibold focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
        />
        <!-- Slug 显示 -->
        <span class="text-sm text-gray-400 hidden md:inline">
          Slug: <code class="bg-gray-700 px-2 py-1 rounded text-blue-400">{{ form.slug || 'auto-generated' }}</code>
        </span>
      </div>

      <!-- 操作按钮 -->
      <div class="flex items-center gap-3 ml-4">
        <!-- 图片上传 -->
        <ImageUploader @upload-success="insertImage" />
        
        <button
          @click="handleSaveDraft"
          :disabled="saving"
          class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition disabled:opacity-50"
        >
          {{ saving ? '保存中...' : '保存草稿' }}
        </button>
        <button
          @click="handlePublish"
          :disabled="publishing"
          class="px-6 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-semibold transition disabled:opacity-50"
        >
          {{ publishing ? (postId ? '更新中...' : '发布中...') : (postId ? '更新文章' : '发布') }}
        </button>
      </div>
    </header>

    <div v-if="loadingPost" class="p-6 text-gray-300 bg-gray-900 border-b border-gray-700">
      正在加载文章内容...
    </div>

    <!-- 主体区域 -->
    <div class="flex-1 flex flex-col md:flex-row overflow-hidden">
      <!-- 左侧：编辑区 -->
      <div class="flex-1 flex flex-col border-r border-gray-700">
        <!-- 分类 & 标签栏 -->
        <div class="bg-gray-800 px-4 py-3 flex flex-wrap gap-4 border-b border-gray-700">
          <!-- 分类选择 -->
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-400">分类:</label>
            <select
              v-model="form.categoryId"
              class="bg-gray-700 border border-gray-600 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option :value="null">无分类</option>
              <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                {{ cat.name }}
              </option>
            </select>
          </div>

          <!-- 标签输入 -->
          <div class="flex items-center gap-2 flex-1">
            <label class="text-sm text-gray-400 whitespace-nowrap">标签:</label>
            <div class="flex flex-wrap gap-2 flex-1">
              <span
                v-for="(tag, index) in form.tags"
                :key="index"
                class="bg-blue-600/20 text-blue-400 px-3 py-1 rounded-full text-sm flex items-center gap-1"
              >
                {{ tag }}
                <button @click="removeTag(index)" class="hover:text-red-400 transition">&times;</button>
              </span>
              <input
                v-model="tagInput"
                @keydown.enter="addTag"
                @keydown.backspace="handleTagBackspace"
                type="text"
                placeholder="输入标签后回车..."
                class="bg-transparent border-none outline-none text-sm flex-1 min-w-[120px]"
              />
            </div>
          </div>
        </div>

        <!-- Markdown 编辑区 -->
        <div class="flex-1 relative">
          <textarea
            v-model="form.content"
            @input="handleContentChange"
            placeholder="开始用 Markdown 写作..."
            class="w-full h-full bg-gray-900 text-gray-100 p-6 resize-none outline-none font-mono text-sm leading-relaxed"
            style="min-height: 500px;"
          ></textarea>
          <!-- 字数统计 -->
          <div class="absolute bottom-4 right-4 text-xs text-gray-500">
            {{ form.content.length }} 字符
          </div>
        </div>
      </div>

      <!-- 右侧：实时预览 -->
      <div class="flex-1 flex flex-col bg-gray-800/50">
        <div class="px-4 py-2 bg-gray-800 border-b border-gray-700 flex items-center justify-between">
          <span class="text-sm text-gray-400 font-semibold">实时预览</span>
          <span class="text-xs text-gray-500">HTML Preview</span>
        </div>
        <div
          v-html="renderedHTML"
          class="flex-1 overflow-y-auto p-6 prose prose-invert prose-blue max-w-none"
        ></div>
      </div>
    </div>

    <!-- 自动保存状态提示 -->
    <Transition name="fade">
      <div
        v-if="autoSaveStatus"
        class="fixed bottom-6 right-6 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 shadow-xl text-sm"
      >
        <span :class="autoSaveStatusColor">{{ autoSaveStatus }}</span>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'
import debounce from 'lodash-es/debounce'
import ImageUploader from './ImageUploader.vue'
import { blogAPI } from '../services/api.js'
import { useRoute } from 'vue-router'

const route = useRoute()
const postId = ref(route.query.id ? parseInt(route.query.id) : null)

// ============================================
// Markdown 渲染器（带代码高亮）
// ============================================
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre class="hljs bg-gray-900 rounded-lg p-4 overflow-x-auto"><code class="hljs language-${lang}">${
          hljs.highlight(str, { language: lang }).value
        }</code></pre>`
      } catch (_) {}
    }
    return `<pre class="hljs bg-gray-900 rounded-lg p-4 overflow-x-auto"><code>${md.utils.escapeHtml(str)}</code></pre>`
  },
})

// ============================================
// 响应式状态
// ============================================
const form = reactive({
  title: '',
  slug: '',
  summary: '',
  content: '',
  categoryId: null,
  tags: [],
})

const tagInput = ref('')
const categories = ref([])
const saving = ref(false)
const publishing = ref(false)
const loadingPost = ref(false)
const autoSaveStatus = ref('')
const autoSaveStatusColor = ref('text-gray-400')

// 渲染后的 HTML（计算属性，实时更新）
const renderedHTML = computed(() => {
  if (!form.content) return '<p class="text-gray-500 italic">在左侧输入 Markdown 内容，这里会实时预览...</p>'
  return md.render(form.content)
})

// ============================================
// 防抖自动保存（30 秒）
// ============================================
const debouncedAutoSave = debounce(async () => {
  if (!form.title && !form.content) return // 空白不保存
  await saveDraft(true) // true 表示自动保存
}, 30000) // 30秒

// 监听内容变化触发防抖
watch(() => form.content, () => {
  debouncedAutoSave()
})

// ============================================
// 方法
// ============================================

// 标题变化 → 自动生成 Slug
const handleTitleChange = () => {
  form.slug = generateSlug(form.title)
}

// 内容变化 → 触发自动保存
const handleContentChange = () => {
  // 已通过 watch 处理
}

// 生成 Slug（简化版）
const generateSlug = (title) => {
  if (!title) return ''
  return title
    .toLowerCase()
    .trim()
    .replace(/[\s]+/g, '-')
    .replace(/[^\w\u4e00-\u9fa5-]/g, '')
    .slice(0, 100)
}

// 添加标签
const addTag = () => {
  const tag = tagInput.value.trim()
  if (tag && !form.tags.includes(tag)) {
    form.tags.push(tag)
  }
  tagInput.value = ''
}

// 移除标签
const removeTag = (index) => {
  form.tags.splice(index, 1)
}

// 处理标签输入框退格键
const handleTagBackspace = () => {
  if (!tagInput.value && form.tags.length > 0) {
    form.tags.pop()
  }
}

// 插入图片到编辑器
const insertImage = (result) => {
  const imageMarkdown = `![${result.name}](${result.url})`
  form.content += '\n' + imageMarkdown + '\n'
  autoSaveStatus.value = '✅ 图片已插入'
  autoSaveStatusColor.value = 'text-green-400'
  setTimeout(() => (autoSaveStatus.value = ''), 2000)
}

// 保存草稿
const saveDraft = async (isAuto = false) => {
  if (!isAuto) saving.value = true
  autoSaveStatus.value = isAuto ? '正在自动保存...' : '正在保存草稿...'
  autoSaveStatusColor.value = 'text-yellow-400'

  try {
    const postData = {
      title: form.title || '无标题',
      slug: form.slug,
      content: form.content,
      category_id: form.categoryId,
      tags: form.tags,
      status: 'draft'
    }
    
    if (postId.value) {
      await blogAPI.updatePost(postId.value, postData)
    } else {
      const result = await blogAPI.createPost(postData)
      if (result?.id) postId.value = result.id
    }
    
    autoSaveStatus.value = isAuto ? '✓ 草稿已自动保存' : '✓ 草稿保存成功'
    autoSaveStatusColor.value = 'text-green-400'

    if (!isAuto) {
      setTimeout(() => (autoSaveStatus.value = ''), 2000)
    }
  } catch (error) {
    console.error('保存失败:', error)
    autoSaveStatus.value = `❌ 保存失败: ${error.message}`
    autoSaveStatusColor.value = 'text-red-400'
  } finally {
    if (!isAuto) saving.value = false
    if (isAuto) {
      setTimeout(() => (autoSaveStatus.value = ''), 3000)
    }
  }
}

// 发布文章
const handleSaveDraft = () => saveDraft(false)

const handlePublish = async () => {
  if (!form.title.trim()) {
    alert('请先输入文章标题')
    return
  }
  publishing.value = true
  try {
    const postData = {
      title: form.title,
      slug: form.slug,
      content: form.content,
      category_id: form.categoryId,
      tags: form.tags,
      status: 'published'
    }
    
    const wasEditing = !!postId.value
    if (postId.value) {
      await blogAPI.updatePost(postId.value, postData)
    } else {
      const result = await blogAPI.createPost(postData)
      if (result?.id) postId.value = result.id
    }
    
    autoSaveStatus.value = wasEditing ? '🎉 文章更新成功！' : '🎉 文章发布成功！'
    autoSaveStatusColor.value = 'text-green-400'
    setTimeout(() => (autoSaveStatus.value = ''), 3000)
  } catch (error) {
    console.error('发布失败:', error)
    alert('发布失败: ' + error.message)
  } finally {
    publishing.value = false
  }
}

// 加载文章内容
const loadPost = async () => {
  if (!postId.value) return
  loadingPost.value = true
  autoSaveStatus.value = '正在加载文章内容...'
  autoSaveStatusColor.value = 'text-yellow-400'
  
  try {
    const post = await blogAPI.getPost(postId.value)
    form.title = post.title || ''
    form.slug = post.slug || ''
    form.content = post.content_markdown || ''
    form.categoryId = post.category_id || null
    form.tags = post.tags ? post.tags.map(t => t.name) : []
    autoSaveStatus.value = '✓ 文章内容已加载，可以修改后更新'
    autoSaveStatusColor.value = 'text-green-400'
    setTimeout(() => (autoSaveStatus.value = ''), 2500)
  } catch (error) {
    console.error('加载文章失败:', error)
    autoSaveStatus.value = `❌ 加载文章失败: ${error.message}`
    autoSaveStatusColor.value = 'text-red-400'
    alert('加载文章失败: ' + error.message)
  } finally {
    loadingPost.value = false
  }
}

// 加载分类
const loadCategories = async () => {
  try {
    const data = await blogAPI.getCategories()
    categories.value = data
  } catch (error) {
    console.warn('⚠️ 加载分类失败，使用默认分类:', error.message)
    // 模拟数据
    categories.value = [
      { id: 1, name: '技术', slug: 'tech' },
      { id: 2, name: '生活', slug: 'life' },
      { id: 3, name: '旅行', slug: 'travel' },
    ]
  }
}

// ============================================
// 生命周期
// ============================================
onMounted(() => {
  loadCategories()
  loadPost()
})

onUnmounted(() => {
  // 清理防抖
  debouncedAutoSave.cancel()
})
</script>

<style scoped>
/* Transition 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
