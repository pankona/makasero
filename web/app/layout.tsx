import type React from "react"
export const metadata = {
  title: "AI Coding Agent",
  description: "An AI-powered coding assistant to help you build applications",
    generator: 'v0.dev'
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}


import './globals.css'