import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'

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
        name='Create Service'
        endpoint='http://localhost:3001/api/services'
        method='POST'
      />
    </div>
  )
}
