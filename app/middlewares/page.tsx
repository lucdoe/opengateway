import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'

export default function Middlewares() {
  return (
    <div>
      <TableWithHeaderAndAddEntry
        name='Middlewares'
        description='Middlewares allow you two extend the Gateway with custom functionality.'
        tableRows={[
          {
            name: 'JWT',
            description: 'Middleware to validate JWT tokens.',
            runsOn: 'Service, Route',
            status: {data: 'Active', useChip: true, color: 'green'},
          },
          {
            name: 'CORS',
            description: 'Middleware to enable CORS.',
            runsOn: 'Global',
            status: {data: 'Active', useChip: true, color: 'green'},
          },
          {
            name: 'Rate Limiting',
            description: 'Middleware to limit the number of requests.',
            runsOn: 'Service, Route',
            status: {data: 'Disabled', useChip: true, color: 'red'},
          },
          {
            name: 'Logging',
            description: 'Middleware to log requests and responses.',
            runsOn: 'Global',
            status: {data: 'Active', useChip: true, color: 'green'},
          },
        ]}
      />
    </div>
  )
}
