import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

export function useAsyncAction() {
  const { t } = useI18n()
  const loading = ref(false)
  const error = ref('')

  async function run(fn: () => Promise<void>) {
    error.value = ''
    loading.value = true
    try {
      await fn()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : t('common.error')
    } finally {
      loading.value = false
    }
  }

  return { loading, error, run }
}
