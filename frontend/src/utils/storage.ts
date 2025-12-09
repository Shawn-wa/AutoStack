/**
 * localStorage 封装
 */
export const storage = {
  /**
   * 获取存储的值
   */
  get<T = string>(key: string): T | null {
    const value = localStorage.getItem(key)
    if (value === null) return null
    
    try {
      return JSON.parse(value) as T
    } catch {
      return value as unknown as T
    }
  },

  /**
   * 设置存储的值
   */
  set(key: string, value: unknown): void {
    if (typeof value === 'string') {
      localStorage.setItem(key, value)
    } else {
      localStorage.setItem(key, JSON.stringify(value))
    }
  },

  /**
   * 移除存储的值
   */
  remove(key: string): void {
    localStorage.removeItem(key)
  },

  /**
   * 清空所有存储
   */
  clear(): void {
    localStorage.clear()
  }
}

/**
 * sessionStorage 封装
 */
export const sessionStore = {
  /**
   * 获取存储的值
   */
  get<T = string>(key: string): T | null {
    const value = sessionStorage.getItem(key)
    if (value === null) return null
    
    try {
      return JSON.parse(value) as T
    } catch {
      return value as unknown as T
    }
  },

  /**
   * 设置存储的值
   */
  set(key: string, value: unknown): void {
    if (typeof value === 'string') {
      sessionStorage.setItem(key, value)
    } else {
      sessionStorage.setItem(key, JSON.stringify(value))
    }
  },

  /**
   * 移除存储的值
   */
  remove(key: string): void {
    sessionStorage.removeItem(key)
  },

  /**
   * 清空所有存储
   */
  clear(): void {
    sessionStorage.clear()
  }
}
