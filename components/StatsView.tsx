import {
  Cog6ToothIcon,
  PuzzlePieceIcon,
  RocketLaunchIcon,
} from '@heroicons/react/24/outline'

const stats = [
  {
    id: 1,
    name: 'Total Services',
    icon: RocketLaunchIcon,
    link: '/services',
  },
  {
    id: 2,
    name: 'Total Endpoints',
    icon: PuzzlePieceIcon,
    link: '/endpoints',
  },
  {
    id: 3,
    name: 'Active Middlewares',
    stat: '3',
    icon: Cog6ToothIcon,
    link: '/middlewares',
  },
]

export default function StatisticsView({
  amountOfServices,
  amountOfEndpoints,
}: {
  amountOfServices: number
  amountOfEndpoints: number
}) {
  return (
    <div className='px-4 sm:px-6 lg:px-8'>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Gateway Statistics
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        See how many total Services, Endpoints and Middlewares you currently
        have in your cluster.
      </p>

      <dl className='mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'>
        {stats.map((item) => (
          <div
            key={item.id}
            className='relative overflow-hidden rounded-lg bg-white px-4 pb-12 pt-5 shadow sm:px-6 sm:pt-6'>
            <dt>
              <div className='absolute rounded-md bg-indigo-500 p-3'>
                <item.icon className='h-6 w-6 text-white' aria-hidden='true' />
              </div>
              <p className='ml-16 truncate text-sm font-medium text-gray-500'>
                {item.name}
              </p>
            </dt>
            <dd className='ml-16 flex items-baseline pb-6 sm:pb-7'>
              <p className='text-2xl font-semibold text-gray-900'>
                {item.stat ||
                  (item.name === 'Total Services' && amountOfServices) ||
                  (item.name === 'Total Endpoints' && amountOfEndpoints) ||
                  0}
              </p>
              <div className='absolute inset-x-0 bottom-0 bg-gray-50 px-4 py-4 sm:px-6'>
                <div className='text-sm'>
                  <a
                    href={item.link}
                    className='font-medium text-indigo-600 hover:text-indigo-500'>
                    View all<span className='sr-only'> {item.name} stats</span>
                  </a>
                </div>
              </div>
            </dd>
          </div>
        ))}
      </dl>
    </div>
  )
}
