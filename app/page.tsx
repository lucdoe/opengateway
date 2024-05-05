'use client'
import DocumentationView from '@/components/DocumentationView'
import StatsView from '@/components/StatsView'
import {apiRoutes} from '@/config/config'
import {useEffect, useState} from 'react'

export default function Dashboard() {
  const [services, setServices] = useState(0)
  const [endpoints, setEndpoints] = useState(0)

  useEffect(() => {
    const fetchServices = async () => {
      const res = await fetch(apiRoutes.services.all)
      const data = await res.json()
      setServices(data.length)
    }

    const fetchEndpoints = async () => {
      const res = await fetch(apiRoutes.endpoints.all)
      const data = await res.json()
      setEndpoints(data.length)
    }

    fetchServices()
    fetchEndpoints()
  }, [])
  return (
    <div className='bg-white'>
      <StatsView amountOfEndpoints={endpoints} amountOfServices={services} />
      <DocumentationView />
    </div>
  )
}
