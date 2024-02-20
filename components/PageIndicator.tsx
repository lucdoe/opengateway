import Link from 'next/link'

export default function PageIndicator({
  page,
  subpage,
}: Readonly<{
  page: string
  subpage: string
}>) {
  return (
    <div className='mb-8'>
      <span className='text-gray-500'>
        <Link href={'/' + page.toLowerCase()}>
          <span className='text-gray-700'>{page}</span>
        </Link>{' '}
        <span className='px-2'>{'>'}</span> {subpage}
      </span>
    </div>
  )
}
