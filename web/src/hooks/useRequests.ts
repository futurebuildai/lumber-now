import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../api/client';
import type { MaterialRequest } from '../types';

export function useRequests(limit = 50, offset = 0) {
  return useQuery({
    queryKey: ['requests', limit, offset],
    queryFn: async () => {
      const { data } = await api.get<{ requests: MaterialRequest[] }>('/requests', {
        params: { limit, offset },
      });
      return data.requests;
    },
  });
}

export function useRequest(id: string) {
  return useQuery({
    queryKey: ['request', id],
    queryFn: async () => {
      const { data } = await api.get<MaterialRequest>(`/requests/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { input_type: string; raw_text?: string; media_url?: string }) => {
      const { data } = await api.post<MaterialRequest>('/requests', body);
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['requests'] }),
  });
}

export function useProcessRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await api.post<MaterialRequest>(`/requests/${id}/process`);
      return data;
    },
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['request', id] });
      qc.invalidateQueries({ queryKey: ['requests'] });
    },
  });
}

export function useConfirmRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, items }: { id: string; items?: unknown[] }) => {
      const { data } = await api.post<MaterialRequest>(`/requests/${id}/confirm`, { items });
      return data;
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['request', id] });
      qc.invalidateQueries({ queryKey: ['requests'] });
    },
  });
}
