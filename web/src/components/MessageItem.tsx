import type { SerializableContent } from "../types"
import { User, Bot } from "lucide-react"
import ReactMarkdown from "react-markdown"
import React from 'react'; // React をインポート

interface MessageItemProps {
  message: SerializableContent
}

// コンポーネント定義を React.memo でラップ
const MessageItem = React.memo(({ message }: MessageItemProps) => {
  // Function to render message content based on part type
  const renderContent = (part: any) => {
    if (part.type === "text") {
      return (
        <div className="prose prose-sm max-w-none">
          <ReactMarkdown>{part.content}</ReactMarkdown>
        </div>
      )
    }

    // Handle other part types (function_call, function_response, etc.)
    return (
      <pre className="bg-gray-100 p-3 rounded-md overflow-x-auto text-xs">{JSON.stringify(part.content, null, 2)}</pre>
    )
  }

  return (
    <div className={`flex ${message.role === "user" ? "justify-end" : "justify-start"}`}>
      <div
        className={`
          max-w-[80%] rounded-lg p-4
          ${message.role === "user" ? "bg-purple-100 text-gray-800" : "bg-white border border-gray-200 shadow-sm"}
        `}
      >
        <div className="flex items-center space-x-2 mb-2">
          {message.role === "user" ? (
            <>
              <span className="font-medium text-purple-700">You</span>
              <User className="h-4 w-4 text-purple-700" />
            </>
          ) : (
            <>
              <Bot className="h-4 w-4 text-gray-700" />
              <span className="font-medium text-gray-700">AI Assistant</span>
            </>
          )}
        </div>

        <div className="space-y-2">
          {message.parts.map((part, index) => (
            <div key={index}>{renderContent(part)}</div>
          ))}
        </div>
      </div>
    </div>
  )
});

// React.memo でラップしたコンポーネントを default export する
export default MessageItem;
