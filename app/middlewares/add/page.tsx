export default function AddMiddleware() {
  return (
    <div>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Add Middleware
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Add a new Middleware to be run by the Gateway. Midllewares can run on a
        service, endpoint or global level.
      </p>
    </div>
  )
}
