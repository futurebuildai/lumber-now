import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../api/client';
import type { InventoryItem } from '../types';

export function useInventory(limit = 50, offset = 0, search?: string) {
  return useQuery({
    queryKey: ['inventory', limit, offset, search],
    queryFn: async () => {
      const { data } = await api.get<{ items: InventoryItem[] }>('/inventory', {
        params: { limit, offset, search },
      });
      return data.items;
    },
  });
}

export function useCreateInventoryItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (item: Partial<InventoryItem>) => {
      const { data } = await api.post<InventoryItem>('/inventory', item);
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['inventory'] }),
  });
}

export function useImportCSV() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData();
      formData.append('file', file);
      const { data } = await api.post('/inventory/import', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['inventory'] }),
  });
}
