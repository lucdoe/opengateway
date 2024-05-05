export default function SubmitBtn({name}: Readonly<{name: string}>) {
  return (
    <button
      type='submit'
      className='mt-8 max-w-48 rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'>
      {name}
    </button>
  )
}
