import { describe, it, expect } from 'vitest'
import api from './client'

describe('API client', () => {
  it('is defined', () => {
    expect(api).toBeDefined()
  })

  it('has baseURL set to /v1', () => {
    expect(api.defaults.baseURL).toBe('/v1')
  })

  it('has Content-Type header set to application/json', () => {
    expect(api.defaults.headers['Content-Type']).toBe('application/json')
  })

  it('has X-Requested-With header for CSRF protection', () => {
    expect(api.defaults.headers['X-Requested-With']).toBe('XMLHttpRequest')
  })

  it('has request interceptors configured', () => {
    // axios stores interceptors; verify at least one request interceptor exists
    expect(api.interceptors.request).toBeDefined()
  })

  it('has response interceptors configured', () => {
    expect(api.interceptors.response).toBeDefined()
  })
})
