import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'

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
        fields={[
          {fieldtype: 'toggle', label: 'Enabled', value: 'true'},
          {fieldtype: 'text', label: 'Name', placeholder: 'Test Name'},

          {fieldtype: 'text', label: 'Path', placeholder: '/test'},
          {
            fieldtype: 'text',
            label: 'Host',
            placeholder: 'localhost',
          },
          {
            fieldtype: 'checkbox',
            label: 'Methods',
            name: 'Methods',
            fields: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
          },
          {
            fieldtype: 'text',
            label: 'Tags',
            placeholder: 'test, example, tag',
          },
        ]}
        name='Create Endpoint'
        endpoint='http://localhost:3001/api/endpoints'
        method='POST'
      />
    </div>
  )
}
