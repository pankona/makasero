"use client"

import { useEffect, useState } from "react"
import dynamic from "next/dynamic"

// Import the Vite app directly
const ViteApp = dynamic(() => import("../src/App"), {
  ssr: false,
})

export default function Page() {
  const [isClient, setIsClient] = useState(false)

  useEffect(() => {
    setIsClient(true)
  }, [])

  if (!isClient) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          flexDirection: "column",
          gap: "1rem",
        }}
      >
        <h1>Loading AI Coding Agent...</h1>
        <div
          style={{
            width: "2rem",
            height: "2rem",
            borderRadius: "50%",
            border: "0.25rem solid #e5e7eb",
            borderTopColor: "#7c3aed",
            animation: "spin 1s linear infinite",
          }}
        ></div>
        <style jsx>{`
          @keyframes spin {
            to {
              transform: rotate(360deg);
            }
          }
        `}</style>
      </div>
    )
  }

  return <ViteApp />
}
