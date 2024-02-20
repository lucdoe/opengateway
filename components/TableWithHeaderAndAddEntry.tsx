'use client'
import {usePathname, useRouter} from 'next/navigation'
import ColoredChip from './ColoredChip'

const generateHeaders = (data: Array<{[key: string]: string}>) => {
  if (data.length > 0) {
    return Object.keys(data[0])
  }
  return []
}

function TableCell({
  content,
}: Readonly<{content: string | {[key: string]: string}}>) {
  if (content && typeof content === 'object' && content.useChip) {
    return (
      <ColoredChip
        color={content.color || ''}
        text={content.data || ''}
        withDot={false}
      />
    )
  } else if (content && typeof content === 'object' && content.useLink) {
    return (
      <ColoredChip
        color={content.color || ''}
        text={content.data || ''}
        withDot
      />
    )
  }

  return <>{content}</>
}

export default function TableWithHeaderAndAddEntry({
  name,
  description,
  tableRows,
}: Readonly<{
  name: string
  description: string
  tableRows: any[]
}>) {
  const columns = generateHeaders(tableRows)
  const pathname = usePathname()
  const router = useRouter()

  const handleRowClick = (rowId: string) => {
    router.push(`http://localhost:3000/${pathname}/${rowId}`)
  }

  const preventPropagation = (e: React.MouseEvent) => {
    e.stopPropagation()
  }

  return (
    <div className='px-4 sm:px-6 lg:px-8'>
      <div className='sm:flex sm:items-center'>
        <div className='sm:flex-auto'>
          <header>
            <h1 className='text-3xl font-bold tracking-tight text-gray-900'>
              {name}
            </h1>
          </header>
          <p className='mt-2 text-sm text-gray-700'>{description}</p>
        </div>
        <div className='mt-4 sm:ml-16 sm:mt-0 sm:flex-none'>
          <button
            type='button'
            className='block rounded-md bg-indigo-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'>
            <a href={'/' + name.toLowerCase() + '/add'}>Add {name}</a>
          </button>
        </div>
      </div>
      <div className='mt-8 flow-root'>
        <div className='-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8'>
          <div className='inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8'>
            <table className='min-w-full divide-y divide-gray-300'>
              <thead>
                <tr>
                  {columns.map((column, index) => {
                    return (
                      <th
                        key={column}
                        scope='col'
                        className={
                          index == 0
                            ? 'py-3.5 pl-3 pr-3 text-left text-sm font-semibold text-gray-900'
                            : 'px-3 pl-3 py-3.5 text-left text-sm font-semibold text-gray-900'
                        }>
                        {column.charAt(0).toUpperCase() + column.slice(1)}
                      </th>
                    )
                  })}

                  <th scope='col' className='relative py-3.5 pl-3 pr-4 sm:pr-0'>
                    <span className='sr-only'>Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className='divide-y divide-gray-200'>
                {tableRows.map((row, rowIndex) => (
                  <tr
                    key={row.id || rowIndex}
                    className='cursor-pointer hover:bg-gray-50 rounded-lg'
                    onClick={() =>
                      handleRowClick(row.id || rowIndex.toString())
                    }>
                    {columns.map((column, index) =>
                      index == 0 ? (
                        <td
                          className='whitespace-nowrap py-4 px-3 text-sm text-gray-700'
                          key={column}>
                          <TableCell content={row[column]} />
                        </td>
                      ) : (
                        <td
                          className='whitespace-nowrap py-4 px-3 text-sm text-gray-500'
                          key={column}>
                          <TableCell content={row[column]} />
                        </td>
                      ),
                    )}
                    <td
                      className='relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0'
                      onClick={preventPropagation}>
                      <a
                        href={`http://localhost:3000/${pathname}/${rowIndex}/edit`}
                        className='text-indigo-600 hover:text-indigo-900 pr-3'
                        onClick={(e) => e.stopPropagation()}>
                        Edit<span className='sr-only'>, {row.name}</span>
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
