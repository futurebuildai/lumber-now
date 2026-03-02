import { describe, it, expect } from 'vitest'
import StatusBadge from './StatusBadge'
import type { RequestStatus } from '@/types'

describe('StatusBadge', () => {
  it('is exported as default', () => {
    expect(StatusBadge).toBeDefined()
    expect(typeof StatusBadge).toBe('function')
  })

  const allStatuses: RequestStatus[] = ['pending', 'processing', 'parsed', 'confirmed', 'sent', 'failed']

  it('all request statuses are handled', () => {
    // StatusBadge should handle every RequestStatus without errors
    for (const status of allStatuses) {
      expect(() => StatusBadge({ status })).not.toThrow()
    }
  })

  it.each(allStatuses)('renders something for status: %s', (status) => {
    const result = StatusBadge({ status })
    expect(result).toBeDefined()
    expect(result).not.toBeNull()
  })

  it('returned element contains the status text', () => {
    const result = StatusBadge({ status: 'pending' })
    // React elements have props.children
    expect(result.props.children).toBe('pending')
  })

  it('returned element has className with status styling', () => {
    const result = StatusBadge({ status: 'failed' })
    expect(result.props.className).toContain('destructive')
  })

  it('returned element has capitalize class', () => {
    const result = StatusBadge({ status: 'pending' })
    expect(result.props.className).toContain('capitalize')
  })

  it('returned element has role prop equal to status', () => {
    const result = StatusBadge({ status: 'pending' })
    expect(result.props.role).toBe('status')
  })

  it('returned element has aria-label containing the status value', () => {
    const result = StatusBadge({ status: 'confirmed' })
    expect(result.props['aria-label']).toContain('confirmed')
  })

  it.each(allStatuses)('returned element has aria-label containing status: %s', (status) => {
    const result = StatusBadge({ status })
    expect(result.props['aria-label']).toContain(status)
  })
})
