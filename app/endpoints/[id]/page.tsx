'use client'
import PageIndicator from '@/components/PageIndicator'
import RenderSingleItem from '@/components/RenderSingleItem'
import SingleItemHeader from '@/components/SingleItemHeader'
import {apiRoutes} from '@/config/config'
import {Endpoint} from '@/interfaces/input'
import {usePathname} from 'next/navigation'
import {useEffect, useState} from 'react'

export default function SingleEndpoint() {
  const [endpoint, setEndpoint] = useState<Endpoint>({
    id: '',
    name: '',
    methods: [''],
    path: '',
    enabled: true,
    tags: [''],
  })
  const pathname = usePathname()

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(apiRoutes.endpoints.one(pathname.split('/')[2]))
      const data = await res.json()
      setEndpoint(data)
    }

    fetchData()
  }, [pathname])

  return (
    <div>
      <PageIndicator page='Endpoints' subpage={endpoint.name} />
      <SingleItemHeader
        category='Endpoint'
        name={endpoint.name}
        text='Page to see the details of a single Endpoint.'
      />
      <RenderSingleItem item={endpoint} />
    </div>
  )
}
