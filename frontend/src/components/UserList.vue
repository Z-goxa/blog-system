<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold">👥 用户管理</h2>
        <p class="text-sm text-gray-400 mt-1">管理访客账号、强制修改密码、调整角色和删除用户</p>
      </div>
      <button @click="loadUsers" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-sm font-semibold transition">
        刷新
      </button>
    </div>

    <div v-if="loading" class="text-center py-20 text-gray-400">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent text-blue-600 rounded-full mb-4" role="status"></div>
      <p>正在加载用户...</p>
    </div>

    <div v-else-if="users.length === 0" class="text-center py-20 bg-gray-800 rounded-xl border border-gray-700">
      <p class="text-gray-400">暂无用户</p>
    </div>

    <div v-else class="bg-gray-800 rounded-xl shadow-lg overflow-hidden border border-gray-700">
      <table class="w-full">
        <thead class="bg-gray-700">
          <tr>
            <th class="px-5 py-3 text-left text-sm">用户</th>
            <th class="px-5 py-3 text-left text-sm">邮箱</th>
            <th class="px-5 py-3 text-left text-sm">角色</th>
            <th class="px-5 py-3 text-left text-sm">状态</th>
            <th class="px-5 py-3 text-left text-sm">注册时间</th>
            <th class="px-5 py-3 text-left text-sm">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-t border-gray-700 hover:bg-gray-700/50 transition">
            <td class="px-5 py-4">
              <div class="font-medium">{{ u.display_name || u.username }}</div>
              <div class="text-xs text-gray-500">@{{ u.username }} · ID {{ u.id }}</div>
            </td>
            <td class="px-5 py-4 text-sm text-gray-300">{{ u.email }}</td>
            <td class="px-5 py-4">
              <select
                :value="u.role"
                @change="handleRoleChange(u, $event.target.value)"
                class="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="subscriber">访客</option>
                <option value="author">作者</option>
                <option value="editor">编辑</option>
                <option value="admin">管理员</option>
              </select>
            </td>
            <td class="px-5 py-4">
              <span :class="statusClass(u.status)" class="px-2 py-1 rounded text-xs">
                {{ statusText(u.status) }}
              </span>
            </td>
            <td class="px-5 py-4 text-sm text-gray-400">{{ formatDate(u.created_at) }}</td>
            <td class="px-5 py-4 whitespace-nowrap">
              <button @click="openPasswordDialog(u)" class="text-yellow-400 hover:text-yellow-300 mr-3 text-sm">
                改密码
              </button>
              <button
                @click="handleDelete(u)"
                :disabled="isCurrentUser(u)"
                :class="isCurrentUser(u) ? 'text-gray-600 cursor-not-allowed' : 'text-red-400 hover:text-red-300'"
                class="text-sm"
              >
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="passwordDialog.user" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div class="w-full max-w-md bg-gray-800 border border-gray-700 rounded-xl p-6 shadow-2xl">
        <h3 class="text-lg font-bold mb-2">强制修改密码</h3>
        <p class="text-sm text-gray-400 mb-4">
          用户：{{ passwordDialog.user.display_name || passwordDialog.user.username }}（@{{ passwordDialog.user.username }}）
        </p>
        <input
          v-model="passwordDialog.password"
          type="password"
          class="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 mb-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="输入新密码，至少 6 位"
          @keyup.enter="handleUpdatePassword"
        >
        <p class="text-xs text-gray-500 mb-5">修改后该用户需要使用新密码重新登录。</p>
        <div class="flex justify-end gap-3">
          <button @click="closePasswordDialog" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition">
            取消
          </button>
          <button @click="handleUpdatePassword" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-sm font-semibold transition">
            确认修改
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { blogAPI } from '../services/api.js'
import { user as currentUser } from '../services/auth.js'

const users = ref([])
const loading = ref(true)
const passwordDialog = ref({ user: null, password: '' })

const loadUsers = async () => {
  loading.value = true
  try {
    users.value = await blogAPI.getUsers()
  } catch (error) {
    console.error('加载用户失败:', error)
    alert('加载用户失败：' + error.message)
    users.value = []
  } finally {
    loading.value = false
  }
}

const isCurrentUser = (u) => Number(currentUser.value?.id) === Number(u.id)

const roleText = (role) => ({
  admin: '管理员',
  editor: '编辑',
  author: '作者',
  subscriber: '访客'
}[role] || role)

const statusText = (status) => ({
  active: '正常',
  inactive: '停用',
  banned: '封禁'
}[status] || status)

const statusClass = (status) => ({
  active: 'bg-green-600/20 text-green-400',
  inactive: 'bg-gray-600/20 text-gray-400',
  banned: 'bg-red-600/20 text-red-400'
}[status] || 'bg-gray-600/20 text-gray-400')

const formatDate = (s) => {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return '-'
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const handleRoleChange = async (u, role) => {
  const oldRole = u.role
  if (!confirm(`确定要把 ${u.display_name || u.username} 的角色从「${roleText(oldRole)}」改为「${roleText(role)}」吗？`)) {
    // 重刷比强行恢复 select 状态简单可靠
    await loadUsers()
    return
  }
  try {
    const updated = await blogAPI.updateUserRole(u.id, role)
    Object.assign(u, updated)
    if (isCurrentUser(u)) {
      const saved = JSON.parse(localStorage.getItem('user') || '{}')
      saved.role = updated.role
      localStorage.setItem('user', JSON.stringify(saved))
      currentUser.value = saved
    }
  } catch (error) {
    console.error('修改角色失败:', error)
    alert('修改角色失败：' + error.message)
    await loadUsers()
  }
}

const openPasswordDialog = (u) => {
  passwordDialog.value = { user: u, password: '' }
}

const closePasswordDialog = () => {
  passwordDialog.value = { user: null, password: '' }
}

const handleUpdatePassword = async () => {
  const target = passwordDialog.value.user
  const password = passwordDialog.value.password
  if (!target) return
  if (!password || password.length < 6) {
    alert('密码长度至少 6 位')
    return
  }
  try {
    await blogAPI.forceUpdateUserPassword(target.id, password)
    alert('密码已修改')
    closePasswordDialog()
  } catch (error) {
    console.error('修改密码失败:', error)
    alert('修改密码失败：' + error.message)
  }
}

const handleDelete = async (u) => {
  if (isCurrentUser(u)) {
    alert('不能删除当前登录账号')
    return
  }
  if (!confirm(`确定删除用户 ${u.display_name || u.username} 吗？该操作不可恢复。`)) return
  try {
    await blogAPI.deleteUser(u.id)
    users.value = users.value.filter(item => item.id !== u.id)
  } catch (error) {
    console.error('删除用户失败:', error)
    alert('删除用户失败：' + error.message)
  }
}

onMounted(loadUsers)
</script>
