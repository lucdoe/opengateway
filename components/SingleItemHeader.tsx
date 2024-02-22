export default function SingleItemHeader({
  category,
  name,
  text,
}: Readonly<{
  category: string
  name: string
  text: string
}>) {
  return (
    <div className='border-b border-gray-200 pb-5 sm:flex sm:items-center sm:justify-between'>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          {category}: <span className='text-gray-500'>{name}</span>
        </h1>
        <p className='mt-2 text-sm text-gray-700'>{text}</p>
      </header>

      <div className='mt-3 sm:ml-4 sm:mt-0'>
        <button
          type='button'
          className='inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'>
          Edit {category}
        </button>
      </div>
    </div>
  )
}
