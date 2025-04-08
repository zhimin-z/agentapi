'use client';

import { useState, FormEvent, KeyboardEvent, useEffect, useRef } from 'react';

interface MessageInputProps {
  onSendMessage: (message: string, type: 'user' | 'raw') => void;
  disabled?: boolean;
}

export default function MessageInput({ onSendMessage, disabled = false }: MessageInputProps) {
  const [message, setMessage] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  
  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSendMessage(message, 'user');
      setMessage('');
    }
  };
  
  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    // If the message is empty, send special keys as raw messages
    if (!message && !disabled) {
      // List of keys to send as raw input when the text field is empty
      const specialKeys: Record<string, string> = {
        'ArrowUp': '\x1b[A',    // Escape sequence for up arrow
        'ArrowDown': '\x1b[B',  // Escape sequence for down arrow
        'ArrowRight': '\x1b[C', // Escape sequence for right arrow
        'ArrowLeft': '\x1b[D',  // Escape sequence for left arrow
        'Escape': '\x1b',       // Escape key
        'Tab': '\t',            // Tab key
        'Delete': '\x1b[3~',    // Delete key
        'Home': '\x1b[H',       // Home key
        'End': '\x1b[F',        // End key
        'PageUp': '\x1b[5~',    // Page Up
        'PageDown': '\x1b[6~',  // Page Down
      };
      
      // Check if the pressed key is in our special keys map
      if (specialKeys[e.key]) {
        e.preventDefault();
        onSendMessage(specialKeys[e.key], 'raw');
        return;
      }
      
      // Handle Enter as raw newline when empty
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        onSendMessage('\r', 'raw');
        return;
      }
      
      // Handle Ctrl+key combinations
      if (e.ctrlKey) {
        const ctrlMappings: Record<string, string> = {
          'c': '\x03', // Ctrl+C (SIGINT)
          'd': '\x04', // Ctrl+D (EOF)
          'z': '\x1A', // Ctrl+Z (SIGTSTP)
          'l': '\x0C', // Ctrl+L (clear screen)
          'a': '\x01', // Ctrl+A (beginning of line) 
          'e': '\x05', // Ctrl+E (end of line)
          'w': '\x17', // Ctrl+W (delete word)
          'u': '\x15', // Ctrl+U (clear line)
        };

        if (ctrlMappings[e.key.toLowerCase()]) {
          e.preventDefault();
          onSendMessage(ctrlMappings[e.key.toLowerCase()], 'raw');
          return;
        }
      }
    } else if (e.key === 'Enter' && !e.shiftKey) {
      // Normal Enter handling for non-empty message
      e.preventDefault();
      handleSubmit(e);
    }
  };
  
  // Keep focus on the textarea for capturing key events
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.focus();
    }
  }, []);
  
  const isRawMode = !message.trim();

  return (
    <form onSubmit={handleSubmit} className="border-t border-gray-300 p-4 bg-white">
      <div className="flex flex-col">
        {isRawMode && !disabled && (
          <div className="mb-1 text-xs text-blue-600 font-mono flex justify-between">
            <span>Raw terminal mode - arrow keys and special keys sent directly</span>
            <span className="text-gray-500">
              Supported: Arrow keys, Tab, Enter, Ctrl+C, Ctrl+D, Ctrl+Z, Ctrl+L, Ctrl+A, Ctrl+E, Ctrl+W, Ctrl+U
            </span>
          </div>
        )}
        <div className="flex">
          <textarea
            ref={textareaRef}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={disabled ? 'Server offline...' : 'Type a message or use arrow keys when empty...'}
            className={`flex-1 resize-none border rounded-l-lg p-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 ${
              isRawMode && !disabled ? 'bg-gray-50 border-blue-200' : 'bg-white'
            }`}
            rows={2}
            disabled={disabled}
          />
          <button
            type="submit"
            disabled={disabled || !message.trim()}
            className="bg-blue-500 text-white px-4 rounded-r-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send
          </button>
        </div>
      </div>
    </form>
  );
}