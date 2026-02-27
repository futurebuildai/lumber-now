import { useMutation } from '@tanstack/react-query'
import api from '@/api/client'
import type { Dealer } from '@/types'
import type { WizardFormData } from '../types'

export function useCreateDealerMutation() {
  return useMutation({
    mutationFn: async (formData: WizardFormData) => {
      const { data } = await api.post<Dealer>('/platform/dealers', {
        name: formData.name,
        slug: formData.slug,
        subdomain: formData.subdomain,
        logo_url: formData.logo_url,
        primary_color: formData.primary_color,
        secondary_color: formData.secondary_color,
        contact_email: formData.contact_email,
        contact_phone: formData.contact_phone,
        address: formData.address,
      })
      return data
    },
  })
}
