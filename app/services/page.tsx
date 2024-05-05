'use client'
import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'
import {apiRoutes} from '@/config/config'
import {useEffect, useState} from 'react'

export default function Services() {
  const [services, setServices] = useState(null)

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(apiRoutes.services.all)
      const data = await res.json()
      setServices(data)
    }

    fetchData()
  }, [])

  return (
    <div>
      <TableWithHeaderAndAddEntry
        name='Services'
        description='List of all services, currently added to your cluster.'
        tableRows={services ?? []}
      />
    </div>
  )
}
