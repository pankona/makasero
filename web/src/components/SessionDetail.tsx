"use client"

import type React from "react"

import { useState, useEffect, useRef } from "react"
import { useParams, Link } from "react-router-dom"
import type { Session } from "../types"
// ↓ Loader2 を追加
import { ArrowLeft, Send, AlertCircle, RefreshCw, Loader2 } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import MessageItem from "./MessageItem"

interface SessionDetailProps {
  onSessionUpdated: () => void
}

export default function SessionDetail({ onSessionUpdated }: SessionDetailProps) {
  const { sessionId } = useParams<{ sessionId: string }>()
  const [session, setSession] = useState<Session | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [command, setCommand] = useState("")
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const prevHistoryRef = useRef<Session["history"] | undefined>(undefined)

  const fetchSession = async () => {
    if (!sessionId) return

    try {
      setLoading(true)
      const response = await fetch(`/api/sessions/${sessionId}`)

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error("Session not found")
        }
        throw new Error("Failed to fetch session")
      }

      const data = await response.json()
      setSession(data)
      setError(null)
    } catch (err) {
      setError(`Failed to load session: ${err instanceof Error ? err.message : "Unknown error"}`)
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const sendCommand = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!sessionId || !command.trim()) return

    try {
      setSending(true)
      const response = await fetch(`/api/sessions/${sessionId}/commands`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command }),
      })

      if (!response.ok) {
        setSending(false);
        throw new Error("Failed to send command")
      }

      // Clear the input
      setCommand("")

      // Refresh the session to get the updated history
      await fetchSession()
      onSessionUpdated()
    } catch (err) {
      setError("Failed to send command. Please try again.")
      console.error(err)
      setSending(false);
    }
  }

  useEffect(() => {
    fetchSession()

    // Set up polling for session updates
    const intervalId = setInterval(fetchSession, 5000) // Poll every 5 seconds

    return () => clearInterval(intervalId)
  }, [sessionId])

  useEffect(() => {
    const currentHistory = session?.history
    const prevHistory = prevHistoryRef.current

    const currentLength = currentHistory?.length ?? 0
    const prevLength = prevHistory?.length ?? 0

    // Scroll to bottom only if the history length has changed and there are messages
    if (currentLength > 0 && currentLength !== prevLength) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
      setSending(false);
    }

    // Update previous history ref
    prevHistoryRef.current = currentHistory
  }, [session?.history])

  if (loading && !session) {
    return (
      <div className="text-center py-12">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-gray-200 border-t-purple-600"></div>
        <p className="mt-4 text-gray-500">Loading session...</p>
      </div>
    )
  }

  if (error && !session) {
    return (
      <div className="max-w-4xl mx-auto">
        <Link to="/" className="inline-flex items-center text-sm text-purple-600 hover:text-purple-700 mb-6">
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Sessions
        </Link>

        <div className="bg-red-50 border border-red-200 rounded-md p-4 flex items-start space-x-3">
          <AlertCircle className="h-5 w-5 text-red-500 mt-0.5" />
          <div>
            <p className="text-red-700">{error}</p>
            <button
              onClick={fetchSession}
              className="mt-2 text-sm text-red-700 hover:text-red-800 font-medium flex items-center"
            >
              <RefreshCw className="h-3 w-3 mr-1" />
              Try Again
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto h-full flex flex-col">
      <Link to="/" className="inline-flex items-center text-sm text-purple-600 hover:text-purple-700 mb-4">
        <ArrowLeft className="h-4 w-4 mr-1" />
        Back to Sessions
      </Link>

      {session && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 flex-1 flex flex-col overflow-hidden">
          <div className="border-b border-gray-200 px-4 py-3 flex justify-between items-center">
            <div>
              <h1 className="text-lg font-semibold text-gray-900">Session {session.id.substring(0, 8)}...</h1>
              <p className="text-xs text-gray-500">Created {formatDistanceToNow(new Date(session.created_at))} ago</p>
            </div>
            <button
              onClick={fetchSession}
              disabled={loading}
              className="flex items-center space-x-1 px-2 py-1 rounded-md text-xs font-medium text-gray-700 hover:bg-gray-100 disabled:opacity-50"
            >
              <RefreshCw className={`h-3 w-3 ${loading ? "animate-spin" : ""}`} />
              <span>Refresh</span>
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {session.history && session.history.length > 0 ? (
              session.history.map((message, index) => <MessageItem key={index} message={message} />)
            ) : (
              <div className="text-center py-8 text-gray-500">No messages yet</div>
            )}
            {sending && (
              <div className="flex items-center justify-center text-gray-500 py-4">
                <Loader2 className="h-5 w-5 mr-2 animate-spin" />
                <span>AI is thinking...</span>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {error && (
            <div className="px-4 py-2 bg-red-50 border-t border-red-200">
              <p className="text-sm text-red-700 flex items-center">
                <AlertCircle className="h-4 w-4 mr-1" />
                {error}
              </p>
            </div>
          )}

          <div className="border-t border-gray-200 p-4">
            <form onSubmit={sendCommand} className="flex space-x-2">
              <input
                type="text"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder="Type your message..."
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-purple-500"
                disabled={sending}
              />
              <button
                type="submit"
                disabled={sending || !command.trim()}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 disabled:opacity-50"
              >
                {sending ? (
                  <span className="inline-block animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"></span>
                ) : (
                  <Send className="h-4 w-4" />
                )}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
