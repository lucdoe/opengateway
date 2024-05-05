import {apiRoutes} from '@/config/config'

export default function MiddlewareCards({
  middlewares,
}: Readonly<{middlewares: any[]}>) {
  const handleAddMiddleware = async (middleware: any) => {
    const res = await fetch(apiRoutes.middlewares.one(middleware.id), {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        enabled: true,
      }),
    })
    if (res.ok) {
      alert(`${middleware.name} has been added to your Gateway.`)
      location.reload()
    }
  }

  return (
    <ul className='grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4'>
      {middlewares.map((middleware) => (
        <li
          key={middleware.name}
          className='col-span-1 flex flex-col divide-y rounded-lg bg-white text-center shadow my-4'>
          <div className='flex flex-1 flex-col p-8'>
            <img
              className='mx-auto h-12 w-12 flex-shrink-0 rounded-full'
              src={middleware.imageUrl}
              alt=''
            />
            <h3 className='mt-6 text-sm font-medium text-gray-900'>
              {middleware.name}
            </h3>
            <dl className='mt-1 flex flex-grow flex-col justify-between'>
              <dt className='sr-only'>Title</dt>
              <dd className='text-sm text-gray-500'>
                {middleware.description}
              </dd>
            </dl>
          </div>
          <div>
            <div className=''>
              <button
                type='button'
                onClick={() => handleAddMiddleware(middleware)}
                className='w-full py-2 text-sm font-medium text-indigo-600 bg-gray-50 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2'>
                + Add Middleware
              </button>
            </div>
          </div>
        </li>
      ))}
    </ul>
  )
}
