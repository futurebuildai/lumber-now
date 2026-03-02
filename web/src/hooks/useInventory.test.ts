import { describe, it, expect } from 'vitest'
import { useInventory, useCreateInventoryItem, useImportCSV } from './useInventory'

describe('useInventory hooks', () => {
  it('exports useInventory', () => {
    expect(useInventory).toBeDefined()
    expect(typeof useInventory).toBe('function')
  })

  it('exports useCreateInventoryItem', () => {
    expect(useCreateInventoryItem).toBeDefined()
    expect(typeof useCreateInventoryItem).toBe('function')
  })

  it('exports useImportCSV', () => {
    expect(useImportCSV).toBeDefined()
    expect(typeof useImportCSV).toBe('function')
  })
})
