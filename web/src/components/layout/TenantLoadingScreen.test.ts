import { describe, it, expect } from 'vitest'
import TenantLoadingScreen from './TenantLoadingScreen'

describe('TenantLoadingScreen', () => {
  it('is exported as default', () => {
    expect(TenantLoadingScreen).toBeDefined()
    expect(typeof TenantLoadingScreen).toBe('function')
  })

  it('renders without throwing', () => {
    expect(() => TenantLoadingScreen()).not.toThrow()
  })

  it('renders loading text', () => {
    const result = TenantLoadingScreen()
    // The inner structure: div > div > [spinner div, p]
    const innerDiv = result.props.children
    const paragraph = innerDiv.props.children[1]
    expect(paragraph.props.children).toBe('Loading...')
  })

  it('has role="status" for accessibility', () => {
    const result = TenantLoadingScreen()
    expect(result.props.role).toBe('status')
  })

  it('has aria-live="polite"', () => {
    const result = TenantLoadingScreen()
    expect(result.props['aria-live']).toBe('polite')
  })

  it('has aria-label="Loading tenant"', () => {
    const result = TenantLoadingScreen()
    expect(result.props['aria-label']).toBe('Loading tenant')
  })

  it('spinner has aria-hidden="true"', () => {
    const result = TenantLoadingScreen()
    // div > div > [spinner, p]
    const innerDiv = result.props.children
    const spinner = innerDiv.props.children[0]
    expect(spinner.props['aria-hidden']).toBe('true')
  })
})
