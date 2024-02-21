'use client'
import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'
import {useEffect, useState} from 'react'

export default function Services() {
  const [services, setServices] = useState(null)

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch('http://localhost:3001/api/services')
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
        tableRows={services ? services : []}
      />
    </div>
  )
}
