export interface Session {
  id: string
  created_at: string
  updated_at: string
  history?: SerializableContent[]
}

export interface SerializableContent {
  role: "user" | "model"
  parts: SerializablePart[]
}

export interface SerializablePart {
  type: string
  content: any
}

export interface CreateSessionRequest {
  prompt: string
}

export interface CreateSessionResponse {
  session_id: string
  status: string
}

export interface SendCommandRequest {
  command: string
}

export interface SendCommandResponse {
  message: string
}
