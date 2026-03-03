import { useParams } from 'react-router-dom'
import CreateDealerWizard from './create-dealer/CreateDealerWizard'

export default function EditDealerPage() {
  const { id } = useParams<{ id: string }>()

  if (!id) return null

  return <CreateDealerWizard mode="edit" dealerId={id} />
}
