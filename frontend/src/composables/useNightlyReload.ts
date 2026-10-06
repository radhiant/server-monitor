import { onMounted, onUnmounted } from 'vue'

/**
 * Triggers a full page reload at a given hour each day.
 *
 * Wall displays run 24/7 and browsers slowly accumulate heap from ECharts
 * canvas instances, DOM nodes, and reactive proxies. A clean reload once a
 * day (default 04:00 AM) keeps the dashboard snappy without human
 * intervention.
 */
export function useNightlyReload(hour = 4) {
  let timer: ReturnType<typeof setInterval> | null = null

  onMounted(() => {
    timer = setInterval(() => {
      const now = new Date()
      if (now.getHours() === hour && now.getMinutes() === 0 && now.getSeconds() < 30) {
        window.location.reload()
      }
    }, 20_000)
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })
}
