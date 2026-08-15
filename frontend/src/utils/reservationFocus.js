export const scrollBehaviorFor = (windowRef = globalThis.window) => {
  const reducedMotion = windowRef?.matchMedia?.(
    '(prefers-reduced-motion: reduce)'
  )?.matches

  return reducedMotion ? 'auto' : 'smooth'
}

export const focusReservationBlock = (element, windowRef = globalThis.window) => {
  if (!element) {
    return false
  }

  element.scrollIntoView({
    behavior: scrollBehaviorFor(windowRef),
    block: 'center',
    inline: 'center'
  })
  element.focus({ preventScroll: true })

  return true
}
