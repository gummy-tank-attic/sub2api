import { describe, expect, it } from 'vitest'
import { formatDisplayVersion } from '../VersionBadge.vue'

describe('formatDisplayVersion', () => {
  it('handles standard semantic versions', () => {
    expect(formatDisplayVersion('0.2.7')).toBe('0.2.7')
    expect(formatDisplayVersion('v0.2.7')).toBe('0.2.7')
    expect(formatDisplayVersion('1.0.0-rc1')).toBe('1.0.0-rc1')
  })

  it('formats full 40-char git commit SHA to 7-char short SHA', () => {
    expect(formatDisplayVersion('30349e156eef0b8a001a75c2822222d805f27c65')).toBe('30349e1')
    expect(formatDisplayVersion('v30349e156eef0b8a001a75c2822222d805f27c65')).toBe('30349e1')
  })

  it('formats full git commit SHA with custom suffix (e.g. -fxvia)', () => {
    expect(formatDisplayVersion('30349e156eef0b8a001a75c2822222d805f27c65-fxvia')).toBe('30349e1-fxvia')
    expect(formatDisplayVersion('v30349e156eef0b8a001a75c2822222d805f27c65-fxvia')).toBe('30349e1-fxvia')
  })

  it('handles empty or blank version input', () => {
    expect(formatDisplayVersion('')).toBe('')
    expect(formatDisplayVersion('   ')).toBe('')
  })
})
