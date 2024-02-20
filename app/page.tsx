import DocumentationView from '@/components/DocumentationView'
import StatsView from '@/components/StatsView'

export default function Dashboard() {
  return (
    <div className='bg-white'>
      <StatsView />
      <DocumentationView />
    </div>
  )
}
