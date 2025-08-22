import './globals.css'
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { Navigation } from '@/components/Navigation'
import { AuthProvider } from '@/components/AuthContext'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'MCP Gateway',
  description: 'Model Context Protocol Gateway Dashboard',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <AuthProvider>
          <Navigation />
          <main style={{ minHeight: '100vh' }}>
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  )
}
