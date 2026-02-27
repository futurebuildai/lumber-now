import { useMutation } from '@tanstack/react-query'
import api from '@/api/client'

interface UploadResult {
  key: string
  url: string
}

export function useUploadLogo() {
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      const { data } = await api.post<UploadResult>('/platform/media/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return data
    },
  })
}
