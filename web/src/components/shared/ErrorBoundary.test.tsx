import { describe, it, expect } from 'vitest'
import ErrorBoundary from './ErrorBoundary'
import { createElement } from 'react'

// Test that ErrorBoundary is exported and is a valid React component
describe('ErrorBoundary', () => {
  it('is defined', () => {
    expect(ErrorBoundary).toBeDefined()
  })

  it('is a class component with getDerivedStateFromError', () => {
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

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function renderError(): any {
    const instance = new ErrorBoundary({ children: null })
    instance.state = { hasError: true, error: new Error('fail') }
    return instance.render()
  }

  it('shows error UI when child throws', () => {
    const child = createElement('p', null, 'Hello world')
    const instance = new ErrorBoundary({ children: child })
    instance.state = ErrorBoundary.getDerivedStateFromError(new Error('Test error'))
    const result = instance.render()
    expect(result).not.toBe(child)
  })

  it('error UI has role="alert"', () => {
    const result = renderError()
    expect(result.props.role).toBe('alert')
  })

  it('error UI has aria-live="assertive"', () => {
    const result = renderError()
    expect(result.props['aria-live']).toBe('assertive')
  })

  it('error UI shows "Something went wrong" heading', () => {
    const result = renderError()
    const heading = result.props.children[0]
    expect(heading.props.children).toBe('Something went wrong')
  })

  it('refresh button has aria-label', () => {
    const result = renderError()
    const button = result.props.children[2]
    expect(button.props['aria-label']).toBe('Refresh page to recover from error')
  })

  it('refresh button displays "Refresh Page" text', () => {
    const result = renderError()
    const button = result.props.children[2]
    expect(button.props.children).toBe('Refresh Page')
  })

  it('initial state has no error', () => {
    const instance = new ErrorBoundary({ children: null })
    expect(instance.state.hasError).toBe(false)
    expect(instance.state.error).toBeNull()
  })
})
