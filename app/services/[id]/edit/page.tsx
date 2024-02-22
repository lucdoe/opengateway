'use client'
import DynamicAddComponent from '@/components/DynamicAddComponent'
import PageIndicator from '@/components/PageIndicator'
import {apiRoutes} from '@/config/config'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'
import {servicesFields} from '../../page'

export default function EditService() {
  const [initialData, setInitialData] = useState<FormData | undefined>()
  const pathname = usePathname()
  const id = pathname.split('/')[2]

  useEffect(() => {
    if (id) {
      fetch(apiRoutes.services.one(id))
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
        .catch((error) => console.error('Failed to load service data', error))
    }
  }, [id])

  return (
    <div>
      <PageIndicator page='Services' subpage='Edit Service' />

      <header>
        <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
          Edit Service
        </h1>
      </header>
      <p className='mt-2 text-sm text-gray-700'>
        Edit your Service in the Gateway.
      </p>
      {initialData && (
        <DynamicAddComponent
          fields={servicesFields as Field[]}
          initialValues={initialData}
          name='Save Service Edit'
          endpoint={apiRoutes.services.one(id)}
          method='PUT'
        />
      )}
    </div>
  )
}
