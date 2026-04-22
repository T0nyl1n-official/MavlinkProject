/**
 * 性能优化工具
 * 实现懒加载、防抖等性能优化功能
 */

/**
 * 防抖函数
 * @param func 要执行的函数
 * @param wait 等待时间（毫秒）
 * @returns 防抖处理后的函数
 */
export function debounce<T extends (...args: any[]) => any>(func: T, wait: number): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  
  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId)
    }
    
    timeoutId = setTimeout(() => {
      func(...args)
      timeoutId = null
    }, wait)
  }
}

/**
 * 节流函数
 * @param func 要执行的函数
 * @param limit 时间限制（毫秒）
 * @returns 节流处理后的函数
 */
export function throttle<T extends (...args: any[]) => any>(func: T, limit: number): (...args: Parameters<T>) => void {
  let inThrottle = false
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      
      setTimeout(() => {
        inThrottle = false
      }, limit)
    }
  }
}

/**
 * 懒加载图片
 * @param options 配置选项
 */
export function lazyLoadImages(options: {
  selector?: string
  threshold?: number
  rootMargin?: string
} = {}) {
  const {
    selector = 'img[data-src]',
    threshold = 0.1,
    rootMargin = '0px 0px 50px 0px'
  } = options
  
  const imageObserver = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement
        const src = img.getAttribute('data-src')
        
        if (src) {
          img.src = src
          img.removeAttribute('data-src')
        }
        
        observer.unobserve(img)
      }
    })
  }, {
    threshold,
    rootMargin
  })
  
  // 观察所有符合条件的图片
  document.querySelectorAll(selector).forEach(img => {
    imageObserver.observe(img)
  })
  
  return imageObserver
}

/**
 * 懒加载组件
 * @param importFn 动态导入函数
 * @returns 懒加载的组件
 */
export function lazyLoadComponent<T>(importFn: () => Promise<{ default: T }>): Promise<{ default: T }> {
  return importFn()
}

/**
 * 优化滚动性能
 * @param callback 滚动回调函数
 * @returns 优化后的滚动处理函数
 */
export function optimizeScroll(callback: () => void): () => void {
  let ticking = false
  
  return () => {
    if (!ticking) {
      requestAnimationFrame(() => {
        callback()
        ticking = false
      })
      ticking = true
    }
  }
}

/**
 * 缓存函数结果
 * @param func 要缓存结果的函数
 * @returns 带缓存的函数
 */
export function memoize<T extends (...args: any[]) => any>(func: T): (...args: Parameters<T>) => ReturnType<T> {
  const cache = new Map<string, ReturnType<T>>()
  
  return (...args: Parameters<T>) => {
    const key = JSON.stringify(args)
    
    if (cache.has(key)) {
      return cache.get(key) as ReturnType<T>
    }
    
    const result = func(...args)
    cache.set(key, result)
    return result
  }
}
