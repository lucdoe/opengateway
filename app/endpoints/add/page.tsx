import DynamicAddComponent from '@/components/DynamicAddComponent'

export default function AddEndpoint() {
  return (
    <div>
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
        ]}
        name='Create Endpoint'
      />
    </div>
  )
}
