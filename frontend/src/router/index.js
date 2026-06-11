import { createRouter, createWebHashHistory } from 'vue-router'

// 后台管理
import Login from '../components/Login.vue'
import AppLayout from '../components/app-layout.vue'
import Dashboard from '../components/Dashboard.vue'
import PostList from '../components/PostList.vue'
import PostEditor from '../components/PostEditor.vue'
import CategoryList from '../components/CategoryList.vue'
import TagList from '../components/TagList.vue'
import CommentList from '../components/CommentList.vue'
import UserList from '../components/UserList.vue'

// 前台博客
import BlogLayout from '../components/blog/BlogLayout.vue'
import BlogHome from '../components/blog/BlogHome.vue'
import BlogPost from '../components/blog/BlogPost.vue'
import BlogAbout from '../components/blog/BlogAbout.vue'

const routes = [
  // 默认跳转博客首页
  { path: '/', redirect: '/blog' },

  // 前台博客（公开访问）
  {
    path: '/blog',
    component: BlogLayout,
    children: [
      { path: '', name: 'BlogHome', component: BlogHome },
      { path: 'about', name: 'BlogAbout', component: BlogAbout },
      { path: 'cat/:cat', name: 'BlogCategory', component: BlogHome }, // 暂复用首页
      { path: ':slug', name: 'BlogPost', component: BlogPost },
    ],
  },

  // 后台管理（需要登录）
  {
    path: '/admin',
    component: AppLayout,
    children: [
      { path: '', name: 'Dashboard', component: Dashboard },
      { path: 'posts', name: 'Posts', component: PostList },
      { path: 'editor', name: 'Editor', component: PostEditor },
      { path: 'categories', name: 'Categories', component: CategoryList },
      { path: 'tags', name: 'Tags', component: TagList },
      { path: 'comments', name: 'Comments', component: CommentList },
      { path: 'users', name: 'Users', component: UserList },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 导航守卫：只在 /admin/* 下才需要登录
const canAccessAdmin = () => {
  const token = localStorage.getItem('token')
  if (!token) return false
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    return user.role && user.role !== 'subscriber'
  } catch {
    return false
  }
}

router.beforeEach((to, from) => {
  const needAuth = to.path.startsWith('/admin')
  const token = localStorage.getItem('token')
  if (needAuth && !token) {
    return '/login'
  }
  if (needAuth && !canAccessAdmin()) {
    return '/blog'
  }
  // 如果已经登录且访问登录页面，只有具备后台权限才重定向到后台
  if (to.path === '/login' && token && canAccessAdmin()) {
    return '/admin'
  }
})

export default router
