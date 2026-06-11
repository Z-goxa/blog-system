<template>
  <div class="blog-shell">
    <!-- 顶部导航条 -->
    <header class="blog-topbar">
      <div class="topbar-inner">
        <div class="brand-row">
          <router-link to="/blog" class="site-title">风雪之隅</router-link>
          <span class="site-slogan">左手代码右手诗</span>
        </div>
        <nav class="topnav">
          <router-link to="/blog">主页</router-link>
          <router-link
            v-for="category in categories"
            :key="category.id"
            :to="`/blog/cat/${category.slug}`"
          >
            {{ category.name }}
          </router-link>
          <router-link to="/blog/about">关于</router-link>
        </nav>
      </div>
    </header>

    <div class="blog-main">
      <!-- 左侧文章内容区 -->
      <main class="blog-content">
        <router-view />
      </main>

      <!-- 右侧 sidebar -->
      <aside class="blog-sidebar">
        <div class="profile">
          <div class="profile-avatar-frame">
            <img src="../../assets/avatar.png" alt="景龙" class="profile-avatar" />
          </div>
          <div class="profile-name">景龙</div>
          <div class="profile-tagline">代码 / 工具 / 思考 / 日常</div>
          <p class="profile-intro">
            这里记录我对技术实践、开源工具和生活观察的整理。把踩过的坑、学到的方法和偶尔冒出来的想法，沉淀成可以回看、也希望能帮到别人的文字。
          </p>
        </div>

        <div class="side-block search-block">
          <h4>按词查询</h4>
          <form class="side-search" @submit.prevent="submitSearch">
            <input v-model="searchPrompt" type="search" placeholder="输入提示词查询文章..." aria-label="提示词查询文章">
            <button type="submit">查询</button>
          </form>
        </div>

        <div class="side-block">
          <h4>归档</h4>
          <ul class="archive-list">
            <li v-if="archives.length === 0" class="archive-empty">暂无归档</li>
            <li v-for="archive in archives" :key="archive.key">
              <router-link :to="archiveLink(archive.key)">
                {{ archive.label }} <span class="archive-count">({{ archive.count }})</span>
              </router-link>
            </li>
          </ul>
        </div>

        <div class="side-block">
          <h4>最新评论</h4>
          <ul class="recent-comment-list">
            <li v-if="recentComments.length === 0" class="archive-empty">暂无评论</li>
            <li v-for="comment in recentComments" :key="comment.id" class="recent-comment-item">
              <router-link :to="`/blog/${comment.post_slug}`" class="comment-link">
                <span class="comment-author">{{ comment.author_name || '匿名用户' }}</span>
                <span class="comment-date">{{ formatCommentDate(comment.created_at) }}</span>
                <span class="comment-content">{{ comment.content }}</span>
                <span class="comment-post">评《{{ comment.post_title }}》</span>
              </router-link>
            </li>
          </ul>
        </div>

        <div class="side-block">
          <h4>友情链接</h4>
          <ul class="archive-list">
            <li><a href="https://www.php.net" target="_blank" rel="noopener">PHP官方</a></li>
            <li><a href="https://github.com/laruence" target="_blank" rel="noopener">Github</a></li>
          </ul>
        </div>
      </aside>
    </div>

    <footer class="blog-footer">
      <div class="footer-inner">
        &copy; 风雪之隅 / 博客声明 / 京ICP备 / PHP 8.1.0-NTS(JIT) / Theme inspired by laruence
      </div>
    </footer>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { API_BASE } from '../../services/api.js'

const route = useRoute()
const router = useRouter()
const categories = ref([])
const archives = ref([])
const recentComments = ref([])
const searchPrompt = ref(String(route.query.q || ''))

const loadCategories = async () => {
  try {
    const response = await fetch(`${API_BASE}/public/categories`)
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || `HTTP ${response.status}`)
    }
    categories.value = await response.json()
  } catch (error) {
    console.error('加载顶部分类导航失败:', error)
    categories.value = []
  }
}

const archiveLink = (key) => ({
  path: route.path.startsWith('/blog/cat/') ? route.path : '/blog',
  query: { ...route.query, archive: key, page: undefined }
})

const submitSearch = () => {
  const q = searchPrompt.value.trim()
  const path = route.path.startsWith('/blog/cat/') ? route.path : '/blog'
  const query = { ...route.query, page: undefined }
  if (q) query.q = q
  else delete query.q
  router.push({ path, query })
}

watch(() => route.query.q, (q) => {
  searchPrompt.value = String(q || '')
})

const formatCommentDate = (s) => {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })
}

const loadArchives = async () => {
  try {
    const response = await fetch(`${API_BASE}/public/archives`)
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || `HTTP ${response.status}`)
    }
    const result = await response.json()
    archives.value = Array.isArray(result) ? result : []
  } catch (error) {
    console.error('加载归档失败:', error)
    archives.value = []
  }
}

const loadRecentComments = async () => {
  try {
    const response = await fetch(`${API_BASE}/public/comments/recent?limit=5`)
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || `HTTP ${response.status}`)
    }
    const result = await response.json()
    recentComments.value = Array.isArray(result) ? result : []
  } catch (error) {
    console.error('加载最新评论失败:', error)
    recentComments.value = []
  }
}

onMounted(() => {
  loadCategories()
  loadArchives()
  loadRecentComments()
})
</script>

<style scoped>
/* ===== 字体 & 全局基调 ===== */
.blog-shell {
  min-height: 100vh;
  background: #fff;
  color: #333;
  font-family: 'Helvetica Neue', 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', sans-serif;
  font-weight: 400;
  font-size: 15px;
  line-height: 1.7;
}

/* ===== 顶部 ===== */
.blog-topbar {
  border-bottom: 1px solid #ededed;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;
}
.topbar-inner {
  max-width: 1140px;
  margin: 0 auto;
  padding: 0 32px;
}
.brand-row {
  display: flex;
  align-items: baseline;
  gap: 16px;
  flex-wrap: wrap;
  padding: 18px 18px 12px;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 45%, #1e40af 100%);
  border-radius: 8px 8px 0 0;
}
.site-title {
  font-size: 26px;
  font-weight: 700;
  color: #fff;
  text-decoration: none;
  letter-spacing: 0.5px;
}
.site-slogan {
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
  font-style: italic;
}
.topnav {
  display: flex;
  gap: 6px;
  font-size: 14px;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 45%, #1e40af 100%);
  border-top: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 0 0 8px 8px;
  padding: 0 10px;
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.22);
  overflow-x: auto;
}
.topnav a {
  color: rgba(255, 255, 255, 0.9);
  text-decoration: none;
  transition: all 0.15s;
  padding: 10px 14px;
  border-bottom: 3px solid transparent;
  white-space: nowrap;
}
.topnav a:hover,
.topnav a.router-link-exact-active {
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
  border-bottom-color: #bfdbfe;
}
.topnav .admin-link {
  color: rgba(255, 255, 255, 0.75);
  font-size: 12px;
}
/* ===== 主体两栏 ===== */
.blog-main {
  max-width: 1140px;
  margin: 0 auto;
  padding: 36px 32px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 40px;
  align-items: start;
}

/* ===== 右 sidebar ===== */
.blog-sidebar {
  font-size: 13px;
  color: #555;
}
.profile {
  border-bottom: 1px solid #ededed;
  padding-bottom: 22px;
  margin-bottom: 24px;
}
.profile::after {
  content: '';
  display: block;
  clear: both;
}
.profile-avatar-frame {
  float: left;
  width: 110px;
  height: 180px;
  border: 1px solid #d9e2f1;
  border-radius: 10px;
  padding: 4px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: 0 8px 22px rgba(37, 99, 235, 0.12);
  overflow: hidden;
  margin-right: 14px;
  margin-bottom: 10px;
}
.profile-avatar {
  width: 100%;
  height: 100%;
  border-radius: 7px;
  object-fit: cover;
  object-position: center top;
  display: block;
}
.profile-name {
  font-weight: 700;
  font-size: 18px;
  color: #1f2937;
  margin-bottom: 4px;
  letter-spacing: 0.4px;
}
.profile-tagline {
  display: inline-block;
  color: #2563eb;
  background: #eff6ff;
  border-radius: 999px;
  padding: 3px 9px;
  font-size: 11.5px;
  line-height: 1.4;
  margin-bottom: 10px;
}
.profile-intro {
  font-size: 12.8px;
  color: #667085;
  line-height: 1.75;
  margin: 0;
  text-align: justify;
  text-indent: 2em;
}
.side-block {
  margin-bottom: 26px;
}
.side-search {
  display: flex;
  gap: 8px;
  align-items: center;
}
.side-search input {
  min-width: 0;
  flex: 1;
  border: 1px solid #dbe4f0;
  border-radius: 999px;
  background: #f8fbff;
  color: #374151;
  outline: none;
  padding: 7px 12px;
  font-size: 12.8px;
}
.side-search input::placeholder {
  color: #9ca3af;
}
.side-search input:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
  background: #fff;
}
.side-search button {
  flex-shrink: 0;
  border: 0;
  border-radius: 999px;
  background: #2563eb;
  color: #fff;
  padding: 7px 13px;
  font-size: 12.8px;
  font-weight: 700;
  cursor: pointer;
}
.side-search button:hover {
  background: #1d4ed8;
}
.side-block h4 {
  font-size: 13px;
  font-weight: 700;
  color: #333;
  margin: 0 0 10px;
  letter-spacing: 0.5px;
}
.archive-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.archive-list li {
  margin-bottom: 6px;
}
.archive-list a {
  color: #666;
  text-decoration: none;
  font-size: 13px;
  font-weight: 300;
}
.archive-empty {
  color: #999;
  font-size: 13px;
}
.archive-count {
  color: #9ca3af;
  font-size: 12px;
}
.archive-list a:hover {
  color: #111;
  text-decoration: underline;
}
.recent-comment-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.recent-comment-item {
  border-bottom: 1px dashed #e5e7eb;
  padding: 8px 0;
}
.recent-comment-item:first-child {
  padding-top: 0;
}
.recent-comment-item:last-child {
  border-bottom: 0;
}
.comment-link {
  display: block;
  color: inherit;
  text-decoration: none;
}
.comment-link:hover .comment-content,
.comment-link:hover .comment-post {
  color: #1d4ed8;
}
.comment-author {
  color: #374151;
  font-weight: 700;
  font-size: 12.5px;
  margin-right: 6px;
}
.comment-date {
  color: #9ca3af;
  font-size: 11.5px;
}
.comment-content {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: #667085;
  font-size: 12.5px;
  line-height: 1.55;
  margin: 3px 0 2px;
}
.comment-post {
  display: block;
  color: #9ca3af;
  font-size: 11.8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ===== 主内容 ===== */
.blog-content {
  min-width: 0; /* grid 子元素防止撑爆 */
}

/* ===== footer ===== */
.blog-footer {
  border-top: 1px solid #ededed;
  background: #fff;
  margin-top: 60px;
}
.footer-inner {
  max-width: 1140px;
  margin: 0 auto;
  padding: 20px 32px;
  font-size: 12px;
  color: #999;
  font-weight: 300;
  text-align: left;
}

/* ===== 响应式 ===== */
@media (max-width: 800px) {
  .blog-main {
    grid-template-columns: 1fr;
    gap: 24px;
    padding: 24px 18px;
  }
  .blog-sidebar {
    border-top: 1px solid #ededed;
    padding-top: 24px;
  }
  .topbar-inner {
    padding: 14px 18px;
  }
  .site-slogan {
    display: none;
  }
  .topnav {
    width: 100%;
    overflow-x: auto;
    gap: 8px;
  }
  .side-search {
    width: 100%;
  }
}
</style>
