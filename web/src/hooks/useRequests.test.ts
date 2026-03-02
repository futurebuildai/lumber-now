import { describe, it, expect } from 'vitest'
import { useRequests, useRequest, useCreateRequest, useProcessRequest, useConfirmRequest } from './useRequests'

describe('useRequests hooks', () => {
  it('exports useRequests', () => {
    expect(useRequests).toBeDefined()
    expect(typeof useRequests).toBe('function')
  })

  it('exports useRequest', () => {
    expect(useRequest).toBeDefined()
    expect(typeof useRequest).toBe('function')
  })

  it('exports useCreateRequest', () => {
    expect(useCreateRequest).toBeDefined()
    expect(typeof useCreateRequest).toBe('function')
  })

  it('exports useProcessRequest', () => {
    expect(useProcessRequest).toBeDefined()
    expect(typeof useProcessRequest).toBe('function')
  })

  it('exports useConfirmRequest', () => {
    expect(useConfirmRequest).toBeDefined()
    expect(typeof useConfirmRequest).toBe('function')
  })
})
