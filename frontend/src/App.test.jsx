import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import App from './App'
import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('App Component', () => {
    beforeEach(() => {
        vi.resetAllMocks()
        
        // Mock scrollIntoView
        window.HTMLElement.prototype.scrollIntoView = vi.fn()
        
        // Default mock implementation for fetch
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/sessions')) {
                return Promise.resolve({
                    ok: true,
                    json: async () => []
                })
            }
            return Promise.reject(new Error(`Unhandled URL: ${url}`))
        })
    })

    it('renders the initial UI', async () => {
        render(<App />)
        expect(screen.getByText('🦁 TeachMe AI')).toBeInTheDocument()
        expect(screen.getByText('Hi! I am your AI study buddy. Ask me anything!')).toBeInTheDocument()
        expect(screen.getByPlaceholderText('Ask a question...')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: 'Send' })).toBeInTheDocument()
        expect(screen.getByText('+ New Chat')).toBeInTheDocument()
    })

    it('allows typing in the input field', async () => {
        render(<App />)
        const input = screen.getByPlaceholderText('Ask a question...')
        fireEvent.change(input, { target: { value: 'Why is the sky blue?' } })
        expect(input.value).toBe('Why is the sky blue?')
    })

    it('sends a message and displays the AI response', async () => {
        const mockResponse = { reply: 'Because of Rayleigh scattering.', session_id: '123' }
        
        // Override mock to handle /chat
        global.fetch.mockImplementation((url, options) => {
            if (url.endsWith('/sessions')) {
                return Promise.resolve({ ok: true, json: async () => [] })
            }
            if (url.endsWith('/chat')) {
                 return Promise.resolve({
                    ok: true,
                    json: async () => mockResponse
                })
            }
             return Promise.reject(new Error(`Unknown URL: ${url}`))
        })

        render(<App />)
        
        const input = screen.getByPlaceholderText('Ask a question...')
        const button = screen.getByRole('button', { name: 'Send' })

        fireEvent.change(input, { target: { value: 'Why is the sky blue?' } })
        fireEvent.click(button)

        // User message should be visible
        expect(screen.getByText('Why is the sky blue?')).toBeInTheDocument()
        
        // Input should be cleared
        expect(input.value).toBe('')
        
        // Loading state should appear
        expect(screen.getByText('Thinking...')).toBeInTheDocument()

        // Wait for AI response
        await waitFor(() => {
            expect(screen.getByText('Because of Rayleigh scattering.')).toBeInTheDocument()
        })
        
        // Verify fetch calls
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/sessions'))
        expect(global.fetch).toHaveBeenCalledWith('http://localhost:8080/chat', expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ message: 'Why is the sky blue?' })
        }))
    })

    it('handles network errors gracefully', async () => {
        global.fetch.mockImplementation((url) => {
            if (url.includes('/sessions')) return Promise.resolve({ ok: true, json: async () => [] })
            if (url.includes('/chat')) return Promise.reject(new Error('Network Error'))
            return Promise.reject(new Error('Unknown'))
        })

        render(<App />)
        
        const input = screen.getByPlaceholderText('Ask a question...')
        const button = screen.getByRole('button', { name: 'Send' })

        fireEvent.change(input, { target: { value: 'Hello' } })
        fireEvent.click(button)

        await waitFor(() => {
             expect(screen.getByText('Sorry, I am having trouble thinking right now.')).toBeInTheDocument()
        })
    })
})
