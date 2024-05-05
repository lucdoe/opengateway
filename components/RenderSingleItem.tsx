import ColoredChip from './ColoredChip'

export default function RenderSingleItem({item}: Readonly<{item: any}>) {
  const handleFieldRender = (key: string): string => {
    if (key === 'enabled') {
      return item[key] ? 'Active' : 'Inactive'
    }

    if (Array.isArray(item[key])) {
      return item[key].join(', ')
    }

    return item[key]
  }

  return (
    <div className='flex justify-between'>
      <div className='w-3/5'>
        {Object.keys(item).map((key: string) => {
          if (key === 'id') return null

          if (key === 'name') {
            return (
              <div
                key={key}
                className='flex justify-between text-gray-800 border-t border-gray-100'>
                <div className='px-4 py-6 sm:px-0'>
                  <dt className='text-base font-medium leading-6 text-gray-900'>
                    {key.charAt(0).toUpperCase() + key.slice(1)}
                  </dt>
                  <dd className='mt-1 text-sm leading-6 text-gray-700'>
                    {handleFieldRender(key)}
                  </dd>
                </div>
                <div className='mr-4 px-4 py-6 sm:px-0'>
                  <dt className='text-base font-medium leading-6 text-gray-900'>
                    Enabled
                  </dt>
                  <dd className='mt-1 text-sm leading-6 text-gray-700'>
                    <ColoredChip
                      text={handleFieldRender('enabled')}
                      color={
                        handleFieldRender('enabled') === 'Active'
                          ? 'green'
                          : 'red'
                      }
                      withDot={true}
                    />
                  </dd>
                </div>
              </div>
            )
          }
          if (key === 'enabled') return null
          return (
            <div key={key} className='text-gray-800 border-t border-gray-100'>
              <dl className='divide-y divide-gray-100'>
                <div className='px-4 py-6 sm:grid sm:grid-cols-2 sm:gap-4 sm:px-0'>
                  <dt className='text-base font-medium leading-6 text-gray-900'>
                    {key.charAt(0).toUpperCase() + key.slice(1)}
                  </dt>
                  <dd className='mt-1 text-sm leading-6 text-gray-700 sm:col-span-2 sm:mt-0'>
                    {handleFieldRender(key)}
                  </dd>
                </div>
              </dl>
            </div>
          )
        })}
      </div>
      <div className='w-2/5 m-8 bg-gray-800 p-6 pl-8 pt-8 rounded-md'>
        <h2 className='text-white font-semibold mb-4'>Endpoint JSON Object</h2>
        <pre className='text-white text-sm'>
          {JSON.stringify(item, null, 4)}
        </pre>
      </div>
    </div>
  )
}
