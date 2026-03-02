import { describe, it, expect } from 'vitest'
import ErrorBoundary from './ErrorBoundary'
import { createElement } from 'react'

// Test that ErrorBoundary is exported and is a valid React component
describe('ErrorBoundary', () => {
  it('is defined', () => {
    expect(ErrorBoundary).toBeDefined()
  })

  it('is a class component with getDerivedStateFromError', () => {
    // ErrorBoundary must be a class component (React limitation)
    expect(typeof ErrorBoundary).toBe('function')
    expect(ErrorBoundary.getDerivedStateFromError).toBeDefined()
  })

  it('getDerivedStateFromError returns error state', () => {
    const error = new Error('test error')
    const state = ErrorBoundary.getDerivedStateFromError(error)
    expect(state).toEqual({ hasError: true, error })
  })

  it('renders children when no error', () => {
    const child = createElement('p', null, 'Hello world')
    const instance = new ErrorBoundary({ children: child })
    instance.state = { hasError: false, error: null }
    const result = instance.render()
    expect(result).toBe(child)
  })

  it('shows error UI when child throws', () => {
    const child = createElement('p', null, 'Hello world')
    const instance = new ErrorBoundary({ children: child })
    // Simulate error being caught via getDerivedStateFromError
    const newState = ErrorBoundary.getDerivedStateFromError(new Error('Test error'))
    instance.state = newState
    const result = instance.render()
    // Should render the error div, not the children
    expect(result).not.toBe(child)
    expect(result.props.role).toBe('alert')
  })

  it('error UI has role="alert"', () => {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    const result = instance.render()
    expect(result.props.role).toBe('alert')
  })

  it('error UI has aria-live="assertive"', () => {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    const result = instance.render()
    expect(result.props['aria-live']).toBe('assertive')
  })

  it('error UI shows "Something went wrong" heading', () => {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    const result = instance.render()
    // div > [h1, p, button]
    const heading = result.props.children[0]
    expect(heading.props.children).toBe('Something went wrong')
  })

  it('refresh button has aria-label', () => {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    const result = instance.render()
    // div > [h1, p, button]
    const button = result.props.children[2]
    expect(button.props['aria-label']).toBe('Refresh page to recover from error')
  })

  it('refresh button displays "Refresh Page" text', () => {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    const result = instance.render()
    // div > [h1, p, button]
    const button = result.props.children[2]
    expect(button.props.children).toBe('Refresh Page')
  })

  it('initial state has no error', () => {
    const instance = new ErrorBoundary({ children: null })
    expect(instance.state.hasError).toBe(false)
    expect(instance.state.error).toBeNull()
  })
})
