import { ref } from 'vue'

export function useTagInput() {
  const values = ref<(string | number)[]>([])
  const tagInputEl = ref<HTMLInputElement | null>(null)

  function parseValue(raw: string): string | number {
    const n = Number(raw)
    return !isNaN(n) && raw.trim() !== '' ? n : raw
  }

  function addTag(e: Event) {
    const input = e.target as HTMLInputElement
    const val = input.value.trim()
    if (val && !values.value.includes(parseValue(val))) {
      values.value.push(parseValue(val))
    }
    input.value = ''
  }

  function removeTag(vi: number) {
    values.value.splice(vi, 1)
  }

  function onBackspace(e: KeyboardEvent) {
    const input = e.target as HTMLInputElement
    if (input.value === '' && values.value.length > 0) values.value.pop()
  }

  function onPaste(e: ClipboardEvent) {
    const pasted = e.clipboardData?.getData('text') ?? ''
    if (!pasted.includes(',') && !pasted.includes('\n')) return
    e.preventDefault()
    const input = e.target as HTMLInputElement
    const parts = (input.value + pasted)
      .split(/[,\n]/)
      .map((s) => s.trim())
      .filter(Boolean)
    for (const part of parts) {
      const parsed = parseValue(part)
      if (!values.value.includes(parsed)) values.value.push(parsed)
    }
    input.value = ''
  }

  return { values, tagInputEl, addTag, removeTag, onBackspace, onPaste }
}
