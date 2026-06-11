<template>
  <div class="blog-post">
    <div v-if="loading" class="state">正在加载...</div>
    <div v-else-if="error" class="state error">{{ error }}</div>

    <article v-else-if="post" class="article">
      <header class="post-head">
        <h1 class="post-title">{{ post.title }}</h1>
        <div class="post-byline">
          Published on
          <a class="date">{{ formatDate(post.published_at || post.created_at) }}</a>
          by <a class="author">laruence</a>
          <span class="dot">·</span>
          <span>{{ post.view_count || 0 }} views</span>
        </div>
      </header>

      <div class="post-content" v-html="renderedContent"></div>

      <footer class="post-foot">
        <div class="post-categories" v-if="post.category">
          <span>Filed in </span>
          <a>{{ post.category.name }}</a>
        </div>
        <div class="post-tags" v-if="post.tags && post.tags.length">
          <span>Tags: </span>
          <a v-for="t in post.tags" :key="t.id">{{ t.name }}</a>
        </div>
      </footer>

      <section class="comments-section">
        <h2>评论</h2>

        <div v-if="commentsLoading" class="comment-state">正在加载评论...</div>
        <div v-else-if="comments.length === 0" class="comment-state">暂无评论，来发表第一条评论吧。</div>
        <div v-else class="comment-list">
          <div v-for="comment in comments" :key="comment.id" class="comment-item">
            <div class="comment-meta">
              <strong>{{ comment.author_name }}</strong>
              <span>{{ formatDate(comment.created_at) }}</span>
            </div>
            <p>{{ comment.content }}</p>
          </div>
        </div>

        <div class="comment-box">
          <div v-if="visitorToken" class="comment-form">
            <div class="visitor-bar">
              当前登录：<strong>{{ visitorUser?.display_name || visitorUser?.username }}</strong>
              <button @click="logoutVisitor">退出</button>
            </div>
            <textarea v-model="commentContent" placeholder="写下你的评论..." maxlength="2000"></textarea>
            <div class="comment-actions">
              <span>{{ commentContent.length }}/2000</span>
              <button :disabled="submittingComment || !commentContent.trim()" @click="submitComment">
                {{ submittingComment ? '提交中...' : '发表评论' }}
              </button>
            </div>
          </div>

          <div v-else class="visitor-auth">
            <h3>{{ authMode === 'login' ? '访客登录后评论' : '注册访客账号' }}</h3>
            <div v-if="authError" class="auth-error">{{ authError }}</div>
            <input v-model="authForm.username" placeholder="用户名" autocomplete="username">
            <input
              v-if="authMode === 'register'"
              v-model="authForm.email"
              placeholder="邮箱"
              autocomplete="email"
            >
            <input
              v-if="authMode === 'register'"
              v-model="authForm.displayName"
              placeholder="昵称（可选）"
              autocomplete="nickname"
            >
            <input v-model="authForm.password" type="password" placeholder="密码（至少 6 位）" autocomplete="current-password">
            <div class="auth-actions">
              <button :disabled="authSubmitting" @click="submitAuth">
                {{ authSubmitting ? '处理中...' : (authMode === 'login' ? '登录' : '注册并登录') }}
              </button>
              <button class="link-btn" @click="toggleAuthMode">
                {{ authMode === 'login' ? '没有账号？注册' : '已有账号？登录' }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <div class="back-link">
        <router-link to="/blog">« 返回主页</router-link>
      </div>
    </article>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { API_BASE } from '../../services/api.js'

const route = useRoute()
const post = ref(null)
const loading = ref(true)
const error = ref('')
const comments = ref([])
const commentsLoading = ref(false)
const commentContent = ref('')
const submittingComment = ref(false)
const visitorToken = ref(localStorage.getItem('visitor_token') || '')
const visitorUser = ref(JSON.parse(localStorage.getItem('visitor_user') || 'null'))
const authMode = ref('login')
const authSubmitting = ref(false)
const authError = ref('')
const authForm = ref({
  username: '',
  email: '',
  displayName: '',
  password: ''
})

const formatDate = (s) => {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric', month: 'long', day: 'numeric'
  })
}

const renderedContent = computed(() => {
  if (!post.value) return ''
  // 后端返回的是 markdown 渲染后的 HTML（content_html）
  return post.value.content_html
    || `<p style="color:#999">（本文暂无正文内容）</p>`
})

const parseError = async (response) => {
  try {
    const data = await response.json()
    return data.error || `HTTP ${response.status}`
  } catch {
    return await response.text() || `HTTP ${response.status}`
  }
}

const loadComments = async (slug) => {
  commentsLoading.value = true
  try {
    const response = await fetch(`${API_BASE}/public/posts/${encodeURIComponent(slug)}/comments`)
    if (!response.ok) throw new Error(await parseError(response))
    comments.value = await response.json()
  } catch (e) {
    console.error('加载评论失败:', e)
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
}

const load = async (slug) => {
  loading.value = true
  error.value = ''
  post.value = null
  try {
    const response = await fetch(`${API_BASE}/public/posts/${encodeURIComponent(slug)}`)
    if (!response.ok) {
      throw new Error(await parseError(response))
    }
    post.value = await response.json()
    await loadComments(slug)
  } catch (e) {
    console.error(e)
    error.value = '加载失败：' + (e?.message || e)
  } finally {
    loading.value = false
  }
}

const saveVisitorSession = (data) => {
  visitorToken.value = data.token
  visitorUser.value = data.user
  localStorage.setItem('visitor_token', data.token)
  localStorage.setItem('visitor_user', JSON.stringify(data.user))
}

const submitAuth = async () => {
  authError.value = ''
  authSubmitting.value = true
  try {
    const url = authMode.value === 'login'
      ? `${API_BASE}/public/auth/login`
      : `${API_BASE}/public/auth/register`
    const body = authMode.value === 'login'
      ? { username: authForm.value.username, password: authForm.value.password }
      : {
          username: authForm.value.username,
          email: authForm.value.email,
          password: authForm.value.password,
          display_name: authForm.value.displayName
        }
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    })
    if (!response.ok) throw new Error(await parseError(response))
    saveVisitorSession(await response.json())
    authForm.value = { username: '', email: '', displayName: '', password: '' }
  } catch (e) {
    authError.value = e?.message || String(e)
  } finally {
    authSubmitting.value = false
  }
}

const toggleAuthMode = () => {
  authMode.value = authMode.value === 'login' ? 'register' : 'login'
  authError.value = ''
}

const logoutVisitor = () => {
  visitorToken.value = ''
  visitorUser.value = null
  localStorage.removeItem('visitor_token')
  localStorage.removeItem('visitor_user')
}

const submitComment = async () => {
  if (!commentContent.value.trim()) return
  submittingComment.value = true
  try {
    const response = await fetch(`${API_BASE}/public/posts/${encodeURIComponent(route.params.slug)}/comments`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${visitorToken.value}`
      },
      body: JSON.stringify({ content: commentContent.value.trim() })
    })
    if (!response.ok) throw new Error(await parseError(response))
    const newComment = await response.json()
    comments.value.push(newComment)
    commentContent.value = ''
  } catch (e) {
    alert('发表评论失败：' + (e?.message || e))
  } finally {
    submittingComment.value = false
  }
}

onMounted(() => load(route.params.slug))
watch(() => route.params.slug, (v) => v && load(v))
</script>

<style scoped>
.state {
  padding: 40px 0;
  color: #999;
  font-size: 14px;
}
.state.error { color: #c33; }

/* ===== 标题区 ===== */
.post-head {
  margin-bottom: 28px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ededed;
}
.post-title {
  margin: 0 0 10px;
  font-size: 30px;
  font-weight: 700;
  color: #222;
  line-height: 1.3;
}
.post-byline {
  font-size: 13px;
  color: #999;
  font-weight: 300;
}
.post-byline a {
  color: #757575;
  text-decoration: none;
}
.post-byline .dot { margin: 0 8px; color: #ccc; }

/* ===== 正文 ===== */
.post-content {
  color: #333;
  font-size: 15px;
  line-height: 1.85;
}

/* :deep 让 v-html 内的标签生效 */
.post-content :deep(h1),
.post-content :deep(h2),
.post-content :deep(h3),
.post-content :deep(h4) {
  font-weight: 700;
  color: #222;
  margin: 1.6em 0 0.6em;
  line-height: 1.4;
}
.post-content :deep(h1) { font-size: 22px; }
.post-content :deep(h2) { font-size: 20px; }
.post-content :deep(h3) { font-size: 17px; }
.post-content :deep(h4) { font-size: 15px; }

.post-content :deep(p) {
  margin: 0.8em 0;
  line-height: 1.85;
}
.post-content :deep(a) {
  color: #2962a3;
  text-decoration: none;
  font-weight: 300;
}
.post-content :deep(a:hover) { text-decoration: underline; }

.post-content :deep(strong),
.post-content :deep(b) { font-weight: 700; color: #222; }
.post-content :deep(em) { font-style: italic; }

/* 引用 */
.post-content :deep(blockquote) {
  background: #f7f7f7;
  border-left: 3px solid #333;
  padding: 8px 16px;
  margin: 1em 0;
  color: #555;
  font-style: italic;
}

/* 代码块 —— 鸟哥同款 */
.post-content :deep(pre) {
  background: #f7f7f7;
  border: 1px solid #ededed;
  border-left: 3px solid #333;
  padding: 12px 14px;
  margin: 1em 0;
  overflow-x: auto;
  font-family: 'Menlo', 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #333;
  border-radius: 0;
}
.post-content :deep(code) {
  background: #f7f7f7;
  padding: 2px 5px;
  font-family: 'Menlo', 'Consolas', 'Monaco', monospace;
  font-size: 0.9em;
  color: #c7254e;
  border: 1px solid #ededed;
}
.post-content :deep(pre code) {
  background: none;
  padding: 0;
  border: none;
  color: inherit;
  font-size: 13px;
}

/* 表格 */
.post-content :deep(table) {
  border-collapse: collapse;
  border-spacing: 0;
  margin: 1em 0;
  font-size: 14px;
}
.post-content :deep(th),
.post-content :deep(td) {
  border: 1px solid #d4d4d4;
  padding: 6px 12px;
}
.post-content :deep(th) {
  background: #f7f7f7;
  font-weight: 700;
}

/* 列表 */
.post-content :deep(ul),
.post-content :deep(ol) {
  padding-left: 1.6em;
  margin: 0.8em 0;
}
.post-content :deep(li) {
  margin: 0.3em 0;
  line-height: 1.7;
}

/* 图片 */
.post-content :deep(img) {
  max-width: 100%;
  height: auto;
  margin: 1em 0;
}

/* hr */
.post-content :deep(hr) {
  border: none;
  border-top: 1px solid #ededed;
  margin: 2em 0;
}

/* ===== 文末 ===== */
.post-foot {
  margin-top: 36px;
  padding-top: 18px;
  border-top: 1px solid #ededed;
  font-size: 13px;
  color: #999;
  display: flex;
  flex-wrap: wrap;
  gap: 22px;
}
.post-foot span { color: #999; }
.post-foot a {
  color: #555;
  text-decoration: none;
  margin-right: 8px;
}
.post-foot a:hover { color: #111; text-decoration: underline; }

/* ===== 评论 ===== */
.comments-section {
  margin-top: 38px;
  padding-top: 24px;
  border-top: 1px solid #ededed;
}
.comments-section h2 {
  font-size: 20px;
  font-weight: 700;
  color: #222;
  margin: 0 0 16px;
}
.comment-state {
  color: #999;
  font-size: 13px;
  padding: 12px 0;
}
.comment-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 20px;
}
.comment-item {
  background: #fafafa;
  border: 1px solid #ededed;
  border-radius: 8px;
  padding: 14px 16px;
}
.comment-meta {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}
.comment-meta strong { color: #333; }
.comment-item p {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.7;
  color: #444;
}
.comment-box {
  margin-top: 20px;
  background: #fff;
  border: 1px solid #ededed;
  border-radius: 10px;
  padding: 16px;
}
.visitor-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #666;
  margin-bottom: 10px;
}
.visitor-bar button,
.link-btn {
  border: 0;
  background: transparent;
  color: #2563eb;
  cursor: pointer;
  font-size: 13px;
}
.visitor-bar button:hover,
.link-btn:hover { text-decoration: underline; }
.comment-form textarea {
  width: 100%;
  min-height: 110px;
  resize: vertical;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 10px 12px;
  outline: none;
  font-family: inherit;
}
.comment-form textarea:focus,
.visitor-auth input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}
.comment-actions,
.auth-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 10px;
}
.comment-actions span {
  font-size: 12px;
  color: #999;
}
.comment-actions button,
.auth-actions button:first-child {
  border: 0;
  border-radius: 8px;
  background: #2563eb;
  color: #fff;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 700;
}
.comment-actions button:disabled,
.auth-actions button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.visitor-auth h3 {
  margin: 0 0 12px;
  font-size: 16px;
  color: #222;
  font-weight: 700;
}
.visitor-auth input {
  width: 100%;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 9px 11px;
  margin-bottom: 10px;
  outline: none;
  font-family: inherit;
}
.auth-error {
  color: #c33;
  font-size: 13px;
  margin-bottom: 10px;
}

.back-link {
  margin-top: 30px;
  font-size: 13px;
}
.back-link a {
  color: #757575;
  text-decoration: none;
  font-weight: 300;
}
.back-link a:hover { color: #111; }
</style>
