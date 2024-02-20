import DynamicAddComponent from '@/components/DynamicAddComponent'

export default function EditEndpoint() {
  return (
    <div>
      <div className='mb-8'>
        <span className='text-gray-500'>
          <span className='text-gray-700'>Endpoints</span>{' '}
          <span className='px-2'>{'>'}</span> Edit Endpoint
        </span>
      </div>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Endpoint
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Edit your Endpoint in the Gateway.
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
        ]}
        name='Save Endpoint Edit'
      />
    </div>
  )
}
