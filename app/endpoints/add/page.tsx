import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {apiRoutes} from '@/config/config'
import {Field, endpointsFields} from '@/interfaces/input'

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
        fields={endpointsFields as Field[]}
        name='Create Endpoint'
        endpoint={apiRoutes.endpoints.all}
        method='POST'
      />
    </div>
  )
}
