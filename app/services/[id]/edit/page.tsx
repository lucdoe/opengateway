import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'

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
        fields={[
          {fieldtype: 'toggle', label: 'Enabled', value: 'true'},
          {},
          {fieldtype: 'text', label: 'Name', placeholder: 'Test Name'},
          {
            fieldtype: 'checkbox',
            label: 'Protocols',
            name: 'Protocols',
            fields: ['http', 'https', 'grpc', 'grpcs'],
          },
          {fieldtype: 'text', label: 'Host', placeholder: 'localhost'},
          {
            fieldtype: 'text',
            label: 'Port',
            placeholder: '8080',
          },

          {fieldtype: 'text', label: 'Tags', placeholder: 'test, tag'},
        ]}
        name='Save Service'
        endpoint='/services'
        method='PUT'
      />
    </div>
  )
}
