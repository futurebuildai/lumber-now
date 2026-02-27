import { useState } from 'react'
import { useInventory, useImportCSV } from '@/hooks/useInventory'
import { Search, Upload, Package, CheckCircle } from 'lucide-react'

export default function InventoryManager() {
  const [search, setSearch] = useState('')
  const { data: items, isLoading } = useInventory(100, 0, search || undefined)
  const importCSV = useImportCSV()

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      await importCSV.mutateAsync(file)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-foreground">Inventory</h1>
        <label className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 cursor-pointer transition-colors">
          <Upload className="h-4 w-4" />
          Import CSV
          <input type="file" accept=".csv" onChange={handleFileUpload} className="hidden" />
        </label>
      </div>

      {importCSV.isSuccess && (
        <div className="flex items-center gap-2 bg-green-500/10 text-green-700 dark:text-green-400 px-4 py-3 rounded-md text-sm">
          <CheckCircle className="h-4 w-4 flex-shrink-0" />
          Import complete: {(importCSV.data as { imported: number }).imported} items imported
        </div>
      )}

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search inventory..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        />
      </div>

      {isLoading ? (
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="flex items-center gap-4 px-6 py-4 border-b border-border last:border-0">
              <div className="h-4 w-24 bg-muted animate-pulse rounded" />
              <div className="h-4 w-40 bg-muted animate-pulse rounded" />
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-4 w-12 bg-muted animate-pulse rounded" />
              <div className="h-4 w-16 bg-muted animate-pulse rounded" />
              <div className="h-5 w-14 bg-muted animate-pulse rounded-full" />
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-card border border-border rounded-lg overflow-hidden">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">SKU</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Category</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Unit</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Price</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">In Stock</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {items?.map((item) => (
                <tr key={item.id} className="hover:bg-muted/50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-foreground">{item.sku}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">{item.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{item.category}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{item.unit}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">${item.price}</td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${item.in_stock ? 'bg-green-500/10 text-green-700 dark:text-green-400' : 'bg-destructive/10 text-destructive'}`}>
                      {item.in_stock ? 'Yes' : 'No'}
                    </span>
                  </td>
                </tr>
              ))}
              {(!items || items.length === 0) && (
                <tr>
                  <td colSpan={6} className="px-6 py-16 text-center">
                    <Package className="h-10 w-10 text-muted-foreground/50 mx-auto mb-3" />
                    <p className="text-sm font-medium text-foreground">No inventory items</p>
                    <p className="text-sm text-muted-foreground mt-1">Import a CSV to get started.</p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
