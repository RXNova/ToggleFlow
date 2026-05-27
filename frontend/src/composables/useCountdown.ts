import { ref, watch, onUnmounted } from 'vue'

export function useCountdown(open: () => boolean, seconds: number) {
  const countdown = ref(0)
  let timer: ReturnType<typeof setInterval> | null = null

  watch(open, (v) => {
    if (v) {
      countdown.value = seconds
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(timer!)
          timer = null
        }
      }, 1000)
    } else {
      if (timer) {
        clearInterval(timer)
        timer = null
      }
    }
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })

  return { countdown }
}
