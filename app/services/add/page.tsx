import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {apiRoutes} from '@/config/config'
import {servicesFields} from '../page'

export default function AddService() {
  return (
    <div>
      <PageIndicator page='Services' subpage='Add Service' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Add Service
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Add a new Service to the Gateway. Services bundle together a set of
        Endpoints.
      </p>
      <DynamicAddComponent
        fields={servicesFields as Field[]}
        name='Create Service'
        endpoint={apiRoutes.services.all}
        method='POST'
      />
    </div>
  )
}
