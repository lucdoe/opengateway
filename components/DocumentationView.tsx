import {BookOpenIcon, CodeBracketIcon} from '@heroicons/react/24/outline'

const items = [
  {
    title: 'Documentation',
    description: 'The source of truth and wisdom.',
    icon: BookOpenIcon,
    background: 'bg-indigo-500',
    href: '/docs',
  },
  {
    title: 'GitHub',
    description: 'Contribute or ask the community for help.',
    icon: CodeBracketIcon,
    background: 'bg-gray-700',
    href: 'github.com/your-github-repo',
  },
]

function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ')
}

export default function DocumentationView() {
  return (
    <div className='mt-12 px-4 sm:px-6 lg:px-8'>
      <h2 className='text-2xl font-bold tracking-tight text-gray-900'>
        Documentation
      </h2>
      <p className='mt-2 text-sm text-gray-700'>
        Here you find more resources and help to configure your Gateway.
      </p>

      <div>
        <ul className='mt-6 grid grid-cols-1 gap-6 border-b border-t border-gray-200 py-6 sm:grid-cols-2'>
          {items.map((item, itemIdx) => (
            <li key={item.title} className='flow-root'>
              <div className='relative -m-2 flex items-center space-x-4 rounded-xl p-2 focus-within:ring-2 focus-within:ring-indigo-500 hover:bg-gray-50'>
                <div
                  className={classNames(
                    item.background,
                    'flex h-16 w-16 flex-shrink-0 items-center justify-center rounded-lg',
                  )}>
                  <item.icon
                    className='h-6 w-6 text-white'
                    aria-hidden='true'
                  />
                </div>
                <div>
                  <h3 className='text-sm font-medium text-gray-900'>
                    <a href={item.href} className='focus:outline-none'>
                      <span className='absolute inset-0' aria-hidden='true' />
                      <span>{item.title}</span>
                      <span aria-hidden='true'> &rarr;</span>
                    </a>
                  </h3>
                  <p className='mt-1 text-sm text-gray-500'>
                    {item.description}
                  </p>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
