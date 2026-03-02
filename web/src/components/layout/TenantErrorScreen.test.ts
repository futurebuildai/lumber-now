import { describe, it, expect } from 'vitest'
import TenantErrorScreen from './TenantErrorScreen'

describe('TenantErrorScreen', () => {
  it('is exported as default', () => {
    expect(TenantErrorScreen).toBeDefined()
    expect(typeof TenantErrorScreen).toBe('function')
  })

  it('renders without throwing', () => {
    expect(() => TenantErrorScreen({ message: 'test' })).not.toThrow()
  })

  it('renders error message passed as prop', () => {
    const result = TenantErrorScreen({ message: 'Something went wrong' })
    // div > div > [icon wrapper, h1, p]
    const innerDiv = result.props.children
    const paragraph = innerDiv.props.children[2]
    expect(paragraph.props.children).toBe('Something went wrong')
  })

  it('has role="alert"', () => {
    const result = TenantErrorScreen({ message: 'error' })
    expect(result.props.role).toBe('alert')
  })

  it('has aria-live="assertive"', () => {
    const result = TenantErrorScreen({ message: 'error' })
    expect(result.props['aria-live']).toBe('assertive')
  })

  it('shows "Tenant Not Found" heading', () => {
    const result = TenantErrorScreen({ message: 'error' })
    // div > div > [icon wrapper, h1, p]
    const innerDiv = result.props.children
    const heading = innerDiv.props.children[1]
    expect(heading.props.children).toBe('Tenant Not Found')
  })

  it('decorative icon container has aria-hidden="true"', () => {
    const result = TenantErrorScreen({ message: 'error' })
    // div > div > [icon wrapper, h1, p]
    const innerDiv = result.props.children
    const iconWrapper = innerDiv.props.children[0]
    expect(iconWrapper.props['aria-hidden']).toBe('true')
  })

  it('svg icon inside container has aria-hidden="true"', () => {
    const result = TenantErrorScreen({ message: 'error' })
    // div > div > [icon wrapper > svg, h1, p]
    const innerDiv = result.props.children
    const iconWrapper = innerDiv.props.children[0]
    const svg = iconWrapper.props.children
    expect(svg.props['aria-hidden']).toBe('true')
  })
})
