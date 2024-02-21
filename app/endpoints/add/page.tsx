import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {endpointsFields} from '../page'

export default function AddEndpoint() {
  return (
    <div>
      <PageIndicator page='Endpoints' subpage='Add Endpoint' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Add Endpoint
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Add a new Endpoint to the Gateway.
      </p>
      <DynamicAddComponent
        fields={endpointsFields}
        name='Create Endpoint'
        endpoint='http://localhost:3001/api/endpoints'
        method='POST'
      />
    </div>
  )
}
