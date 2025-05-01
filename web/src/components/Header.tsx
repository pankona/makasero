"use client"

import { Link, useLocation } from "react-router-dom"
import { PlusCircle, Code, Menu, X } from "lucide-react"
import { useState } from "react"

export default function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const location = useLocation()

  const toggleMobileMenu = () => {
    setMobileMenuOpen(!mobileMenuOpen)
  }

  return (
    <header className="bg-white shadow-sm">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center py-4">
          <Link to="/" className="flex items-center space-x-2">
            <Code className="h-6 w-6 text-purple-600" />
            <span className="text-xl font-bold text-gray-900">AI Coding Agent</span>
          </Link>

          {/* Desktop navigation */}
          <nav className="hidden md:flex items-center space-x-4">
            <Link
              to="/"
              className={`px-3 py-2 rounded-md text-sm font-medium ${
                location.pathname === "/" ? "bg-purple-100 text-purple-700" : "text-gray-700 hover:bg-gray-100"
              }`}
            >
              Sessions
            </Link>
            <Link
              to="/sessions/new"
              className="flex items-center space-x-1 px-3 py-2 rounded-md text-sm font-medium bg-purple-600 text-white hover:bg-purple-700"
            >
              <PlusCircle className="h-4 w-4" />
              <span>New Session</span>
            </Link>
          </nav>

          {/* Mobile menu button */}
          <button className="md:hidden p-2 rounded-md text-gray-700 hover:bg-gray-100" onClick={toggleMobileMenu}>
            {mobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>

        {/* Mobile navigation */}
        {mobileMenuOpen && (
          <div className="md:hidden py-2 border-t border-gray-200">
            <Link
              to="/"
              className={`block px-3 py-2 rounded-md text-base font-medium ${
                location.pathname === "/" ? "bg-purple-100 text-purple-700" : "text-gray-700 hover:bg-gray-100"
              }`}
              onClick={() => setMobileMenuOpen(false)}
            >
              Sessions
            </Link>
            <Link
              to="/sessions/new"
              className="flex items-center space-x-1 px-3 py-2 rounded-md text-base font-medium text-purple-600 hover:bg-gray-100"
              onClick={() => setMobileMenuOpen(false)}
            >
              <PlusCircle className="h-4 w-4" />
              <span>New Session</span>
            </Link>
          </div>
        )}
      </div>
    </header>
  )
}
