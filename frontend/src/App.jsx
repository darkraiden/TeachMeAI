import React, { useState, useEffect, useRef } from 'react'
import './App.css'

function App() {
    const [messages, setMessages] = useState([
        { role: 'ai', text: 'Hi! I am your AI study buddy. Ask me anything!' }
    ])
    const [input, setInput] = useState('')
    const [loading, setLoading] = useState(false)
    const [sessionId, setSessionId] = useState(null)
    const [sessions, setSessions] = useState([])
    const [editingSessionId, setEditingSessionId] = useState(null)
    const [editTitle, setEditTitle] = useState('')
    const messagesEndRef = useRef(null)

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
    }

    useEffect(() => {
        scrollToBottom()
    }, [messages])

    useEffect(() => {
        fetchSessions()
    }, [sessionId]) 

    const fetchSessions = async () => {
        try {
            const res = await fetch('http://localhost:8080/sessions')
            if (res.ok) {
                const data = await res.json()
                setSessions(data || [])
            }
        } catch (err) {
            console.error("Failed to fetch sessions", err)
        }
    }

    const loadSession = async (id) => {
        if (id === sessionId) return;
        try {
            const res = await fetch(`http://localhost:8080/sessions?id=${id}`)
            if (res.ok) {
                const data = await res.json()
                setSessionId(data.id)
                // Normalize 'assistant' -> 'ai' for UI
                const msgs = data.messages ? data.messages.map(m => ({
                    role: m.role === 'assistant' ? 'ai' : m.role,
                    text: m.content
                })) : []
                
                if (msgs.length === 0) {
                     setMessages([{ role: 'ai', text: 'Hi! I am your AI study buddy. Ask me anything!' }])
                } else {
                    setMessages(msgs)
                }
            }
        } catch (err) {
            console.error("Failed to load session", err)
        }
    }

    const deleteSession = async (e, id) => {
        e.stopPropagation()
        if (!window.confirm("Are you sure you want to delete this conversation?")) return

        try {
            const res = await fetch(`http://localhost:8080/sessions?id=${id}`, {
                method: 'DELETE'
            })
            if (res.ok) {
                if (sessionId === id) {
                    createNewChat()
                }
                fetchSessions()
            }
        } catch (err) {
            console.error("Failed to delete session", err)
        }
    }

    const startEditing = (e, session) => {
        e.stopPropagation()
        setEditingSessionId(session.id)
        setEditTitle(session.title || "")
    }

    const saveTitle = async (e, id) => {
        e.stopPropagation()
        try {
            const res = await fetch('http://localhost:8080/sessions', {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id, title: editTitle })
            })
            if (res.ok) {
                setEditingSessionId(null)
                fetchSessions()
            }
        } catch (err) {
            console.error("Failed to update title", err)
        }
    }

    const createNewChat = () => {
        setSessionId(null)
        setMessages([{ role: 'ai', text: 'Hi! I am your AI study buddy. Ask me anything!' }])
    }

    const sendMessage = async (e) => {
        e.preventDefault()
        if (!input.trim()) return

        const userMsg = { role: 'user', text: input }
        setMessages(prev => [...prev, userMsg])
        const currentInput = input
        setInput('')
        setLoading(true)

        try {
            const payload = { message: currentInput }
            if (sessionId) {
                payload.session_id = sessionId
            }

            const response = await fetch('http://localhost:8080/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            })

            if (!response.ok) {
                throw new Error('Network response was not ok')
            }

            const data = await response.json()

            // Store session ID for conversation continuity
            if (data.session_id && !sessionId) {
                setSessionId(data.session_id)
            }

            const aiMsg = { role: 'ai', text: data.reply }
            setMessages(prev => [...prev, aiMsg])
        } catch (error) {
            console.error('Error:', error)
            setMessages(prev => [...prev, { role: 'ai', text: 'Sorry, I am having trouble thinking right now.' }])
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="app-layout">
            <aside className="sidebar">
                <button className="new-chat-btn" onClick={createNewChat}>
                    + New Chat
                </button>
                <div className="session-list">
                    {sessions.map(session => (
                        <div 
                            key={session.id} 
                            className={`session-item ${session.id === sessionId ? 'active' : ''}`}
                            onClick={() => loadSession(session.id)}
                        >
                            <div className="session-header">
                                <div className="session-date">
                                    {new Date(session.updated_at).toLocaleDateString()}
                                </div>
                                <div className="session-actions">
                                    <button className="icon-btn edit-btn" onClick={(e) => startEditing(e, session)} title="Rename">✎</button>
                                    <button className="icon-btn delete-btn" onClick={(e) => deleteSession(e, session.id)} title="Delete">🗑️</button>
                                </div>
                            </div>
                            
                            {editingSessionId === session.id ? (
                                <div className="edit-area" onClick={e => e.stopPropagation()}>
                                    <input 
                                        type="text" 
                                        value={editTitle} 
                                        onChange={e => setEditTitle(e.target.value)}
                                        className="edit-input"
                                        autoFocus
                                    />
                                    <button className="save-btn" onClick={(e) => saveTitle(e, session.id)}>✓</button>
                                </div>
                            ) : (
                                <div className="session-preview">
                                    {session.title || (session.messages && session.messages.length > 0 
                                        ? session.messages[session.messages.length - 1].content 
                                        : 'New Conversation')}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </aside>

            <main className="main-content">
                <header>
                    <h1>🦁 TeachMe AI</h1>
                </header>

                <div className="chat-window">
                    {messages.map((msg, idx) => (
                        <div key={idx} className={`message ${msg.role}`}>
                            <div className="bubble">
                                {msg.text}
                            </div>
                        </div>
                    ))}
                    {loading && <div className="message ai"><div className="bubble">Thinking...</div></div>}
                    <div ref={messagesEndRef} />
                </div>

                <form onSubmit={sendMessage} className="input-area">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder="Ask a question..."
                    />
                    <button type="submit" className="send-btn" disabled={loading}>Send</button>
                </form>
            </main>
        </div>
    )
}

export default App
