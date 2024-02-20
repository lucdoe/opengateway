import DarkOverlapNav from '@/components/DarkOverlapNav'
import type {Metadata} from 'next'
import {Inter} from 'next/font/google'
import './globals.css'

const inter = Inter({subsets: ['latin']})

export const metadata: Metadata = {
  title: 'Capstone Gateway GUI',
  description: 'A GUI for managing the Capstone Gateway.',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang='en' className='h-full bg-white'>
      <body className={inter.className + 'h-full'}>
        <div className='min-h-full bg-white'>
          <DarkOverlapNav />
          <main className='-mt-32'>
            <div className='mx-auto max-w-screen-xl px-4 pb-12 sm:px-6 lg:px-8'>
              <div className='rounded-lg bg-white px-5 py-6 shadow sm:px-6'>
                {children}
              </div>
            </div>
          </main>
        </div>
      </body>
    </html>
  )
}
