'use client'
import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'
import {apiRoutes} from '@/config/config'
import {useEffect, useState} from 'react'

export default function Middlewares() {
  const [middlewares, setMiddlewares] = useState([])

  useEffect(() => {
    const fetchData = async () => {
      const res = await fetch(apiRoutes.middlewares.all)
      let data = await res.json()
      data = data.filter((middleware: any) => middleware.enabled)
      setMiddlewares(data)
    }

    fetchData()
  }, [])

  return (
    <div>
      <TableWithHeaderAndAddEntry
        name='Middlewares'
        description='Middlewares allow you two extend the Gateway with custom functionality.'
        tableRows={middlewares ?? []}
      />
    </div>
  )
}
