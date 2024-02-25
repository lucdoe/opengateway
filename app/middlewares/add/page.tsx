'use client'
import MiddlewareCards from '@/components/MiddlewareCards'
import {apiRoutes} from '@/config/config'
import {useEffect, useState} from 'react'

export default function AddMiddleware() {
  const [middlewares, setMiddlewares] = useState([])

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(apiRoutes.middlewares.all)
      let data = await res.json()
      data = data.filter((middleware: any) => middleware.enabled === false)
      setMiddlewares(data)
    }

    fetchData()
  }, [])

  return (
    <div>
      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Add Middleware
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        See available middlewares and choose one to add to your Gateway.
      </p>
      <MiddlewareCards middlewares={middlewares} />
    </div>
  )
}
