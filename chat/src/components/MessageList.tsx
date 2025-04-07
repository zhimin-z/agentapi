'use client';

import { useEffect, useRef } from 'react';

interface Message {
  role: string;
  content: string;
}

interface MessageListProps {
  messages: Message[];
}

export default function MessageList({ messages }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  // Scroll to bottom when messages change
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);
  
  // If no messages, show a placeholder
  if (messages.length === 0) {
    return (
      <div className="flex-1 p-6 flex items-center justify-center text-gray-500 bg-white">
        <p>No messages yet. Start the conversation!</p>
      </div>
    );
  }
  
  return (
    <div className="flex-1 overflow-y-auto p-4 bg-white">
      {messages.map((message, index) => (
        <div key={index} className={`mb-4 ${message.role === 'user' ? 'text-right' : ''}`}>
          <div
            className={`inline-block max-w-[80%] px-4 py-2 rounded-lg ${
              message.role === 'user'
                ? 'bg-blue-500 text-white rounded-tr-none'
                : 'bg-gray-200 text-gray-800 rounded-tl-none'
            }`}
          >
            <div className="text-xs mb-1 font-bold">
              {message.role === 'user' ? 'You' : 'OpenAgent'}
            </div>
            <div className="whitespace-pre-wrap break-words">{message.content}</div>
          </div>
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
}