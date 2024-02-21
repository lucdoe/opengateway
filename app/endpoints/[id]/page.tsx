'use client'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'

export default function SingleEndpoint() {
  const [service, setService] = useState(null)
  const pathname = usePathname()

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(
        `http://localhost:3001/api/services/${pathname.split('/').pop()}`,
      )
      const data = await res.json()
      setService(data)
    }

    fetchData()
  }, [])

  return (
    <div>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Single Endpoint
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Page to see all the details of a single endpoint.
      </p>
      <div className='mt-4'>
        <pre className='text-gray-800'>{JSON.stringify(service, null, 2)}</pre>
      </div>
    </div>
  )
}
