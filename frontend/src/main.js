import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'

// 捕获全局错误
window.onerror = function(message, source, lineno, colno, error) {
  console.error('🔥 全局错误:', message, 'at', source, ':', lineno, ':', colno, error)
  const appDiv = document.getElementById('app')
  if (appDiv && appDiv.innerHTML === '') {
    appDiv.innerHTML = `<div style="padding: 20px; color: white; background: #991b1b; border-radius: 8px; margin: 20px;">
      <h2 style="margin-top: 0">💥 启动错误</h2>
      <p>${message}</p>
      <pre style="font-size: 12px; opacity: 0.8">${error?.stack || ''}</pre>
    </div>`
  }
}

try {
  const app = createApp(App)
  app.use(router)
  
  // 处理路由错误
  router.onError((error) => {
    console.error('🚀 路由错误:', error)
  })

  app.mount('#app')
} catch (err) {
  console.error('❌ 应用挂载失败:', err)
}
