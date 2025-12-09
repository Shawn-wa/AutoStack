/**
 * 格式化日期时间
 */
export function formatDateTime(dateStr: string | Date, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!dateStr) return '-'
  
  const date = typeof dateStr === 'string' ? new Date(dateStr) : dateStr
  
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 格式化日期
 */
export function formatDate(dateStr: string | Date): string {
  return formatDateTime(dateStr, 'YYYY-MM-DD')
}

/**
 * 格式化时间
 */
export function formatTime(dateStr: string | Date): string {
  return formatDateTime(dateStr, 'HH:mm:ss')
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${units[i]}`
}

/**
 * 格式化数字（千分位）
 */
export function formatNumber(num: number): string {
  return num.toLocaleString('zh-CN')
}

/**
 * 截断文本
 */
export function truncate(text: string, length: number, suffix = '...'): string {
  if (!text || text.length <= length) return text
  return text.slice(0, length) + suffix
}
