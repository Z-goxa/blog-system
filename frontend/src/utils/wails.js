import { ref } from 'vue'

export const isWailsReady = ref(!!(window.go && window.go.main))

export const isWails = () => {
  const hasGo = !!(window.go && window.go.main)
  const hasBlogService = !!(hasGo && window.go.main.BlogService)
  const hasRuntime = !!window.runtime

  return hasBlogService || (hasGo && hasRuntime)
}

export const waitForWails = (timeout = 10000, force = false) => {
  if (force) {
    isWailsReady.value = false
  }

  return new Promise((resolve) => {
    if (isWails()) {
      isWailsReady.value = true
      resolve(true)
      return
    }

    const startTime = Date.now()
    const check = () => {
      if (isWails()) {
        isWailsReady.value = true
        resolve(true)
      } else if (Date.now() - startTime > timeout) {
        console.warn('Wails 初始化超时')
        resolve(false)
      } else {
        setTimeout(check, 100)
      }
    }
    check()
  })
}

setInterval(() => {
  if (!isWailsReady.value && isWails()) {
    isWailsReady.value = true
  }
}, 1000)

waitForWails()
