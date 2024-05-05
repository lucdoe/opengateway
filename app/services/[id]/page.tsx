'use client'
import RenderSingleItem from '@/components/RenderSingleItem'
import SingleItemHeader from '@/components/SingleItemHeader'
import {apiRoutes} from '@/config/config'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'

interface Service {
  id: string
  name: string
  protocols: [string]
  host: string
  port: string
  enabled: boolean
  tags: [string]
}
export default function SingleService() {
  const [service, setService] = useState<Service>({
    id: '',
    name: '',
    protocols: [''],
    host: '',
    port: '',
    enabled: true,
    tags: [''],
  })
  const pathname = usePathname()

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(apiRoutes.services.one(pathname.split('/')[2]))
      const data = await res.json()
      setService(data)
    }

    fetchData()
  }, [pathname])

  return (
    <div>
      <SingleItemHeader
        category='Service'
        name={service.name}
        text='Page to see the details of a single Service.'
        id={service.id}
      />
      <RenderSingleItem item={service} />
    </div>
  )
}
