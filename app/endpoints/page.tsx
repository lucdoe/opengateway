import TableWithHeaderAndAddEntry from '@/components/TableWithHeaderAndAddEntry'

export default function Endpoints() {
  return (
    <div>
      <TableWithHeaderAndAddEntry
        name='Endpoints'
        description='List of all endpoints, currently added to your cluster.'
        tableRows={[
          {
            name: 'Test/ Dev',
            methods: 'GET, POST, PUT, PATCH, DELETE',
            paths: '/test, /dev',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'test, dev',
          },
          {
            name: 'User Service',
            methods: 'GET, POST, PUT, PATCH',
            paths: '/user',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Finance Service',
            methods: 'GET, POST',
            paths: '/finance',
            status: {data: 'Disabled', useChip: true, color: 'red'},
            tags: 'prod',
          },
          {
            name: 'Delivery Service',
            methods: 'GET, POST, PUT, DELETE',
            paths: '/delivery',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Order Service',
            methods: 'GET, POST, PUT, PATCH',
            paths: '/order',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Payment Service',
            methods: 'GET, POST',
            paths: '/payment',
            status: {data: 'Disabled', useChip: true, color: 'red'},
            tags: 'prod',
          },
          {
            name: 'Notification Service',
            methods: 'GET, POST, PUT, PATCH',
            paths: '/notification',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Feedback Service',
            methods: 'GET, POST',
            paths: '/feedback',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
          {
            name: 'Support Service',
            methods: 'GET, POST, PUT, PATCH',
            paths: '/support',
            status: {data: 'Active', useChip: true, color: 'green'},
            tags: 'prod',
          },
        ]}
      />
    </div>
  )
}
