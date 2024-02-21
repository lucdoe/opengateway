import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {servicesFields} from '../../page'

export default function EditService() {
  return (
    <div>
      <PageIndicator page='Services' subpage='Edit Service' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Service
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Edit your Service in the Gateway.
      </p>
      <DynamicAddComponent
        fields={servicesFields}
        name='Save Service'
        endpoint='http://localhost:3001/api/services'
        method='PUT'
      />
    </div>
  )
}
