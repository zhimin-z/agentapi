'use client';

import { useState, useEffect, useRef } from 'react';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import { useSearchParams } from 'next/navigation'

interface Message {
  role: string;
  content: string;
  id: number;
}

interface MessageUpdateEvent {
  id: number;
  role: string;
  message: string;
  time: string;
}

interface StatusChangeEvent {
  status: string;
}

export default function ChatInterface() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [serverStatus, setServerStatus] = useState<string>('unknown');
  const searchParams = useSearchParams();
  // null port gets converted to NaN
  const parsedPort = parseInt(searchParams.get('port') as string);
  const port = isNaN(parsedPort) ? 3284 : parsedPort;
  const openAgentUrl = `http://localhost:${port}`;
  const eventSourceRef = useRef<EventSource | null>(null);

  // Set up SSE connection to the events endpoint
  useEffect(() => {
    // Function to create and set up EventSource
    const setupEventSource = () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      
      const eventSource = new EventSource(`${openAgentUrl}/events`);
      eventSourceRef.current = eventSource;
      
      // Handle message updates
      eventSource.addEventListener('message_update', (event) => {
        const data: MessageUpdateEvent = JSON.parse(event.data);
        
        setMessages(prevMessages => {
          // Check if message with this ID already exists
          const existingIndex = prevMessages.findIndex(m => m.id === data.id);
          
          if (existingIndex !== -1) {
            // Update existing message
            const updatedMessages = [...prevMessages];
            updatedMessages[existingIndex] = {
              role: data.role,
              content: data.message,
              id: data.id
            };
            return updatedMessages;
          } else {
            // Add new message
            return [...prevMessages, {
              role: data.role,
              content: data.message,
              id: data.id
            }];
          }
        });
      });
      
      // Handle status changes
      eventSource.addEventListener('status_change', (event) => {
        const data: StatusChangeEvent = JSON.parse(event.data);
        setServerStatus(data.status);
      });
      
      // Handle connection open (server is online)
      eventSource.onopen = () => {
        // Connection is established, but we'll wait for status_change event
        // for the actual server status
      };
      
      // Handle connection errors
      eventSource.onerror = (error) => {
        console.error('EventSource error:', error);
        setServerStatus('offline');
        
        // Try to reconnect after delay
        setTimeout(() => {
          if (eventSourceRef.current) {
            setupEventSource();
          }
        }, 3000);
      };
      
      return eventSource;
    };
    
    // Initial setup
    const eventSource = setupEventSource();
    
    // Clean up on component unmount
    return () => {
      eventSource.close();
    };
  }, [openAgentUrl]);
  
  // Send a new message
  const sendMessage = async (content: string, type: 'user' | 'raw' = 'user') => {
    // For user messages, require non-empty content
    if (type === 'user' && !content.trim()) return;
    
    // For raw messages, don't set loading state as it's usually fast
    if (type === 'user') {
      setLoading(true);
    }
    
    try {
      const response = await fetch(`${openAgentUrl}/message`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          content, 
          type 
        }),
      });
      
      if (!response.ok) {
        console.error('Failed to send message:', await response.json());
      }
    } catch (error) {
      console.error('Error sending message:', error);
    } finally {
      if (type === 'user') {
        setLoading(false);
      }
    }
  };
  
  return (
    <div className="flex flex-col h-[80vh] bg-gray-100 rounded-lg overflow-hidden border border-gray-300 shadow-lg w-full max-w-[95vw]">
      <div className="p-3 bg-gray-800 text-white text-sm flex justify-between items-center">
        <span>OpenAgent Chat</span>
        <span className="flex items-center">
          <span className={`w-2 h-2 rounded-full mr-2 ${["offline", "unknown"].includes(serverStatus) ? 'bg-red-500' : 'bg-green-500'}`}></span>
          <span>Status: {serverStatus}</span>
          <span className="ml-2">Port: {port}</span>
        </span>
      </div>
      
      <MessageList messages={messages} />
      
      <MessageInput onSendMessage={sendMessage} disabled={loading} />
    </div>
  );
}