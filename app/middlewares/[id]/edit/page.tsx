'use client'
import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {apiRoutes} from '@/config/config'
import {Field, middlewaresFields} from '@/interfaces/input'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'

export default function EditMiddleware() {
  const [initialData, setInitialData] = useState()
  const pathname = usePathname()
  const id = pathname.split('/')[2]

  useEffect(() => {
    if (id) {
      fetch(apiRoutes.middlewares.one(id))
        .then((response) => response.json())
        .then((data) => {
          setInitialData({
            ...data,
          })
        })
        .catch((error) => console.error('Failed to load endpoint data', error))
    }
  }, [id])

  return (
    <div>
      <PageIndicator page='Middlewares' subpage='Edit Middleware' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Middleware Configuration
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Allows you to edit your Middleware Configuration in the Gateway.
      </p>
      {initialData && (
        <DynamicAddComponent
          fields={middlewaresFields as Field[]}
          initialValues={initialData}
          name='Save Middleware Config'
          endpoint={apiRoutes.middlewares.one(id)}
          method='PUT'
        />
      )}
    </div>
  )
}
