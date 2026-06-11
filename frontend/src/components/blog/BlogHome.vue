<template>
  <div class="blog-home">
    <div v-if="activeSearch || activeArchive" class="search-summary">
      <template v-if="activeSearch">
        正在按提示词查询：<strong>{{ activeSearch }}</strong>
      </template>
      <template v-if="activeArchive">
        <span v-if="activeSearch" class="summary-separator">·</span>
        正在查看归档：<strong>{{ activeArchiveLabel }}</strong>
      </template>
      <router-link :to="activeCategory ? `/blog/cat/${activeCategory}` : '/blog'">清除筛选</router-link>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>正在加载文章...</p>
    </div>
    
    <div v-else-if="error" class="error-container">
      <div class="error-icon">⚠️</div>
      <p>{{ error }}</p>
    </div>
    
    <div v-else-if="posts.length === 0" class="empty-container">
      <div class="empty-icon">📝</div>
      <p>暂无文章，敬请期待！</p>
    </div>

    <!-- 文章列表 -->
    <div class="posts-grid" v-else>
      <article v-for="post in posts" :key="post.id" class="post-card">
        <div class="post-card-inner">
          <!-- 文章图片：只有文章确实配置了封面图时才显示；无图时不再显示空白占位块 -->
          <div class="post-image" v-if="coverImage(post)">
            <img :src="coverImage(post)" :alt="post.title" @error="hideBrokenImage">
          </div>

          <div class="post-content">
            <header class="post-header">
              <div class="post-category" v-if="post.category">
                <span class="category-tag">{{ post.category.name }}</span>
              </div>
              <h2 class="post-title">
                <router-link :to="`/blog/${post.slug}`">{{ post.title }}</router-link>
              </h2>
            </header>

            <div class="post-byline">
              <span class="author">laruence</span>
              <span class="separator">·</span>
              <span class="date">{{ formatDate(post.published_at || post.created_at) }}</span>
              <span class="separator">·</span>
              <span class="views">{{ post.view_count || 0 }} 阅读</span>
            </div>

            <div class="post-excerpt" v-html="renderExcerpt(post)"></div>

            <div class="post-footer">
              <div class="post-tags" v-if="post.tags && post.tags.length">
                <span class="tag" v-for="t in post.tags" :key="t.id">
                  {{ t.name }}
                </span>
              </div>
              <router-link :to="`/blog/${post.slug}`" class="read-more">
                阅读更多
                <span class="arrow">→</span>
              </router-link>
            </div>
          </div>
        </div>
      </article>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="totalPages > 1">
      <button :disabled="page <= 1" @click="goPage(page - 1)" class="page-btn">
        <span class="arrow">←</span>
        上一页
      </button>
      <div class="page-info">
        <span class="current-page">{{ page }}</span>
        <span class="separator">/</span>
        <span class="total-pages">{{ totalPages }}</span>
      </div>
      <button :disabled="page >= totalPages" @click="goPage(page + 1)" class="page-btn">
        下一页
        <span class="arrow">→</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { API_BASE } from '../../services/api.js'

const route = useRoute()
const posts = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 8
const loading = ref(true)
const error = ref('')

const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / pageSize))
)

const activeCategory = computed(() => route.params.cat ? String(route.params.cat) : '')
const activeSearch = computed(() => route.query.q ? String(route.query.q).trim() : '')
const activeArchive = computed(() => route.query.archive ? String(route.query.archive).trim() : '')
const activeArchiveLabel = computed(() => {
  if (!activeArchive.value) return ''
  const [year, month] = activeArchive.value.split('-')
  return `${year}年${Number(month)}月`
})

const formatDate = (s) => {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric', month: 'long', day: 'numeric'
  })
}

const renderExcerpt = (post) => {
  if (post.excerpt) return escapeHtml(post.excerpt)
  return '<span style="color:#999">（无摘要）</span>'
}

const coverImage = (post) => {
  const url = post?.meta?.cover_image
  return typeof url === 'string' && url.trim() ? url.trim() : ''
}

const hideBrokenImage = (event) => {
  const container = event?.currentTarget?.closest?.('.post-image')
  if (container) container.style.display = 'none'
}

const escapeHtml = (s) =>
  String(s || '')
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      page_size: String(pageSize)
    })
    if (activeCategory.value) params.set('category', activeCategory.value)
    if (activeSearch.value) params.set('q', activeSearch.value)
    if (activeArchive.value) params.set('archive', activeArchive.value)
    const response = await fetch(`${API_BASE}/public/posts?${params.toString()}`)
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || `HTTP ${response.status}`)
    }
    const result = await response.json()
    posts.value = Array.isArray(result.posts) ? result.posts : []
    total.value = Number(result.total) || posts.value.length
  } catch (e) {
    console.error(e)
    error.value = '加载失败：' + (e?.message || e)
    posts.value = []
  } finally {
    loading.value = false
  }
}

const goPage = (p) => {
  page.value = p
  window.scrollTo({ top: 0, behavior: 'smooth' })
  load()
}

onMounted(load)

watch(() => [route.params.cat, route.query.q, route.query.archive], async () => {
  page.value = 1
  await load()
})
</script>

<style scoped>
/* 全局容器 */
.blog-home {
  min-height: 100vh;
  background: #fafafa;
}

.search-summary {
  max-width: 1160px;
  margin: 0 auto 8px;
  padding: 12px 18px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 10px;
  color: #1e3a8a;
  font-size: 14px;
}
.search-summary strong { color: #1d4ed8; }
.summary-separator {
  margin: 0 6px;
  color: #93c5fd;
}
.search-summary a {
  margin-left: 12px;
  color: #2563eb;
  font-weight: 700;
  text-decoration: none;
}
.search-summary a:hover { text-decoration: underline; }

/* 加载状态 */
.loading-container,
.error-container,
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  text-align: center;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #e0e0e0;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-icon,
.empty-icon {
  font-size: 4rem;
  margin-bottom: 20px;
}

.error-container p {
  color: #c33;
  font-size: 1.1rem;
}

.empty-container p {
  color: #999;
  font-size: 1.1rem;
}

/* 文章网格 */
.posts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 30px;
  padding: 40px 20px;
  max-width: 1200px;
  margin: 0 auto;
}

/* 文章卡片 */
.post-card {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  cursor: pointer;
}

.post-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.12);
}

.post-card-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 文章图片 */
.post-image {
  width: 100%;
  height: 200px;
  overflow: hidden;
}

.post-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.post-card:hover .post-image img {
  transform: scale(1.05);
}

/* 文章内容 */
.post-content {
  padding: 25px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.post-header {
  margin-bottom: 15px;
}

.post-category {
  margin-bottom: 10px;
}

.category-tag {
  background: #667eea;
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.post-title {
  font-size: 1.3rem;
  font-weight: 700;
  line-height: 1.4;
  margin: 0 0 10px;
  color: #222;
}

.post-title a {
  color: inherit;
  text-decoration: none;
  transition: color 0.3s ease;
}

.post-title a:hover {
  color: #667eea;
}

.post-byline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #999;
  margin-bottom: 15px;
}

.post-byline .author {
  font-weight: 500;
  color: #667eea;
}

.post-byline .separator {
  color: #e0e0e0;
}

.post-byline .views {
  font-weight: 300;
}

.post-excerpt {
  color: #666;
  font-size: 0.95rem;
  line-height: 1.6;
  margin-bottom: 20px;
  flex: 1;
}

/* 文章标签 */
.post-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-top: auto;
}

.post-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  flex: 1;
}

.tag {
  background: #f5f7fa;
  color: #667eea;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  transition: all 0.3s ease;
}

.tag:hover {
  background: #667eea;
  color: white;
}

.read-more {
  color: #667eea;
  font-size: 0.9rem;
  font-weight: 500;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.3s ease;
}

.read-more:hover {
  color: #764ba2;
  gap: 8px;
}

/* 分页 */
.pagination {
  display: flex;
  gap: 15px;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

.page-btn {
  background: white;
  border: 1px solid #e0e0e0;
  padding: 10px 20px;
  font-size: 1rem;
  color: #666;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.3s ease;
  font-family: inherit;
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn:hover:not(:disabled) {
  background: #667eea;
  color: white;
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
}

.page-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
  color: #666;
  font-weight: 500;
}

.current-page {
  color: #667eea;
  font-weight: 700;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-subtitle {
    font-size: 1.2rem;
  }
  
  .posts-grid {
    grid-template-columns: 1fr;
    gap: 20px;
    padding: 30px 15px;
  }
  
  .post-content {
    padding: 20px;
  }
  
  .pagination {
    flex-direction: column;
    gap: 10px;
  }
}
</style>
