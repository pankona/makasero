"use client"

import { useState, useEffect } from "react"
import { BrowserRouter as Router, Routes, Route } from "react-router-dom"
import SessionList from "./components/SessionList"
import SessionDetail from "./components/SessionDetail"
import NewSession from "./components/NewSession"
import Header from "./components/Header"
import type { Session } from "./types"
import "./index.css"

function App() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchSessions = async () => {
    try {
      setLoading(true)
      const response = await fetch("http://localhost:8080/api/sessions")
      if (!response.ok) {
        throw new Error("Failed to fetch sessions")
      }
      const data = await response.json()
      // Sort sessions by updated_at in descending order
      const sortedSessions = data.sort((a: Session, b: Session) => 
        new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
      );
      setSessions(sortedSessions)
      setError(null)
    } catch (err) {
      setError("Failed to load sessions. Please try again later.")
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchSessions()

    // Set up polling for session updates
    const intervalId = setInterval(fetchSessions, 10000) // Poll every 10 seconds

    return () => clearInterval(intervalId)
  }, [])

  return (
    <Router>
      <div className="min-h-screen bg-gray-50 flex flex-col">
        <Header />
        <main className="flex-1 container mx-auto px-4 py-6">
          <Routes>
            <Route
              path="/"
              element={<SessionList sessions={sessions} loading={loading} error={error} onRefresh={fetchSessions} />}
            />
            <Route path="/sessions/new" element={<NewSession onSessionCreated={fetchSessions} />} />
            <Route path="/sessions/:sessionId" element={<SessionDetail onSessionUpdated={fetchSessions} />} />
          </Routes>
        </main>
        <footer className="bg-white border-t border-gray-200 py-4">
          <div className="container mx-auto px-4 text-center text-gray-500 text-sm">
            © {new Date().getFullYear()} AI Coding Agent
          </div>
        </footer>
      </div>
    </Router>
  )
}

export default App
