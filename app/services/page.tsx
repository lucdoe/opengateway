'use client'
import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'
import {useEffect, useState} from 'react'

export const servicesFields = [
  {fieldtype: 'toggle', label: 'Enabled', value: 'true'},
  {},
  {fieldtype: 'text', label: 'Name', placeholder: 'Test Name'},
  {
    fieldtype: 'checkbox',
    label: 'Protocols',
    name: 'Protocols',
    fields: ['http', 'https', 'grpc', 'grpcs'],
  },
  {fieldtype: 'text', label: 'Host', placeholder: 'localhost'},
  {
    fieldtype: 'text',
    label: 'Port',
    placeholder: '8080',
  },

  {fieldtype: 'text', label: 'Tags', placeholder: 'test, tag'},
]

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
