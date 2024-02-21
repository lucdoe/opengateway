import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {endpointsFields} from '../../page'

export default function EditEndpoint() {
  return (
    <div>
      <PageIndicator page='Endpoints' subpage='Edit Endpoint' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Endpoint
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Edit your Endpoint in the Gateway.
      </p>
      <DynamicAddComponent
        fields={endpointsFields}
        name='Save Endpoint Edit'
        endpoint='http://localhost:3001/api/endpoints'
        method='PUT'
      />
    </div>
  )
}
