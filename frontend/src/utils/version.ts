/**
 * Format raw version strings for display.
 * Compresses full 40-character Git commit SHAs into standard 7-character short SHAs,
 * preserving any custom branch/build suffixes (e.g. 30349e156eef...-fxvia -> 30349e1-fxvia).
 */
export function formatDisplayVersion(rawVersion: string): string {
  if (!rawVersion) return ''
  const trimmed = rawVersion.trim()
  const stripped = trimmed.replace(/^v/i, '')

  const fullShaMatch = stripped.match(/^([0-9a-fA-F]{7})[0-9a-fA-F]{20,}(.*)$/)
  if (fullShaMatch) {
    return `${fullShaMatch[1]}${fullShaMatch[2]}`
  }

  return stripped
}
