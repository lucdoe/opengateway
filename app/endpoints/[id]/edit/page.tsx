'use client'
import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {apiRoutes} from '@/config/config'
import {Field, endpointsFields} from '@/interfaces/input'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'

export default function EditEndpoint() {
  const [initialData, setInitialData] = useState<FormData | undefined>()
  const pathname = usePathname()
  const id = pathname.split('/')[2]

  useEffect(() => {
    if (id) {
      fetch(apiRoutes.endpoints.one(id))
        .then((response) => response.json())
        .then((data) => {
          const methodsObject = data.methods.reduce(
            (acc: {[x: string]: boolean}, method: string | number) => {
              acc[method] = true
              return acc
            },
            {},
          )

          setInitialData({
            ...data,
            methods: methodsObject,
          })
        })
        .catch((error) => console.error('Failed to load endpoint data', error))
    }
  }, [id])

  return (
    <div>
      <PageIndicator page='Endpoints' subpage='Edit Endpoint' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Endpoint
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Edit your Endpoint in the Gateway.
      </p>
      {initialData && (
        <DynamicAddComponent
          fields={endpointsFields as Field[]} 
          initialValues={initialData}
          name='Save Endpoint Edit'
          endpoint={apiRoutes.endpoints.one(id)}
          method='PUT'
        />
      )}
    </div>
  )
}
