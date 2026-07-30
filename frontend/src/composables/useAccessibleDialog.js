import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const selector = 'a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])'

export const useAccessibleDialog = ({ visible, close, focusOnOpen = true }) => {
  const dialogRef = ref(null)
  let previousFocus = null
  const focusables = () => Array.from(dialogRef.value?.querySelectorAll(selector) || [])
  const onKeydown = event => {
    if (event.key === 'Escape') {
      event.preventDefault()
      close()
      return
    }
    if (event.key !== 'Tab') return
    const items = focusables()
    if (!items.length) {
      event.preventDefault()
      dialogRef.value?.focus()
      return
    }
    const [first] = items
    const last = items.at(-1)
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }
  watch(visible, async open => {
    if (open) {
      previousFocus = document.activeElement
      await nextTick()
      if (typeof focusOnOpen === 'boolean' ? focusOnOpen : focusOnOpen.value) {
        ;(focusables()[0] || dialogRef.value)?.focus()
      }
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
      previousFocus?.focus?.()
      previousFocus = null
    }
  }, { immediate: true })
  onBeforeUnmount(() => {
    document.body.style.overflow = ''
    previousFocus?.focus?.()
  })
  return { dialogRef, onKeydown }
}
