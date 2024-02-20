import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'

export default function Services() {
  return (
    <div>
      <TableWithHeaderAndAddEntry
        name='Services'
        description='List of all services, currently added to your cluster.'
        tableRows={[
          {
            name: 'Test/ Dev',
            protocol: 'http, https',
            host: 'localhost',
            port: '3000',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'test, dev',
          },
          {
            name: 'User Service',
            protocol: 'https',
            host: 'domain.com',
            port: '-',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Finance Service',
            protocol: 'https, grpc',
            host: 'domain.com',
            port: '-',
            status: {data: 'Disabled', useChip: true, color: 'red'},
            tags: 'prod',
          },
          {
            name: 'Delivery Service',
            protocol: 'https, grpc',
            host: 'domain.com',
            port: '-',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
        ]}
      />
    </div>
  )
}
