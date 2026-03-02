import { describe, it, expect } from 'vitest'
import { hexToRgb, hexToHsl, contrastForeground, lighten } from './color-utils'

describe('hexToRgb', () => {
  it('converts valid hex with hash', () => {
    expect(hexToRgb('#ff0000')).toEqual({ r: 255, g: 0, b: 0 })
  })

  it('converts valid hex without hash', () => {
    expect(hexToRgb('00ff00')).toEqual({ r: 0, g: 255, b: 0 })
  })

  it('converts black', () => {
    expect(hexToRgb('#000000')).toEqual({ r: 0, g: 0, b: 0 })
  })

  it('converts white', () => {
    expect(hexToRgb('#ffffff')).toEqual({ r: 255, g: 255, b: 255 })
  })

  it('handles mixed case', () => {
    expect(hexToRgb('#FF00ff')).toEqual({ r: 255, g: 0, b: 255 })
  })

  it('returns null for invalid hex', () => {
    expect(hexToRgb('not-a-hex')).toBeNull()
  })

  it('returns null for short hex', () => {
    expect(hexToRgb('#fff')).toBeNull()
  })

  it('returns null for empty string', () => {
    expect(hexToRgb('')).toBeNull()
  })
})

describe('hexToHsl', () => {
  it('converts red', () => {
    const hsl = hexToHsl('#ff0000')
    expect(hsl).toEqual({ h: 0, s: 100, l: 50 })
  })

  it('converts white to achromatic', () => {
    const hsl = hexToHsl('#ffffff')
    expect(hsl).toEqual({ h: 0, s: 0, l: 100 })
  })

  it('converts black to achromatic', () => {
    const hsl = hexToHsl('#000000')
    expect(hsl).toEqual({ h: 0, s: 0, l: 0 })
  })

  it('converts green', () => {
    const hsl = hexToHsl('#00ff00')
    expect(hsl).toEqual({ h: 120, s: 100, l: 50 })
  })

  it('converts blue', () => {
    const hsl = hexToHsl('#0000ff')
    expect(hsl).toEqual({ h: 240, s: 100, l: 50 })
  })

  it('returns null for invalid input', () => {
    expect(hexToHsl('invalid')).toBeNull()
  })
})

describe('contrastForeground', () => {
  it('returns black for white background', () => {
    expect(contrastForeground('#ffffff')).toBe('#000000')
  })

  it('returns white for black background', () => {
    expect(contrastForeground('#000000')).toBe('#ffffff')
  })

  it('returns white for dark blue', () => {
    expect(contrastForeground('#000080')).toBe('#ffffff')
  })

  it('returns black for yellow', () => {
    expect(contrastForeground('#ffff00')).toBe('#000000')
  })

  it('returns white for invalid hex', () => {
    expect(contrastForeground('invalid')).toBe('#ffffff')
  })
})

describe('lighten', () => {
  it('lightens black by 50%', () => {
    const result = lighten('#000000', 0.5)
    expect(result).toBe('#808080')
  })

  it('does not change white', () => {
    const result = lighten('#ffffff', 0.5)
    expect(result).toBe('#ffffff')
  })

  it('lightens by 0 returns same color', () => {
    const result = lighten('#ff0000', 0)
    expect(result).toBe('#ff0000')
  })

  it('lightens by 1 returns white', () => {
    const result = lighten('#000000', 1)
    expect(result).toBe('#ffffff')
  })

  it('returns input for invalid hex', () => {
    expect(lighten('invalid', 0.5)).toBe('invalid')
  })
})
