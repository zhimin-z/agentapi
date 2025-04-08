'use client';

import { useState, useEffect } from 'react';
import MessageList from './MessageList';
import MessageInput from './MessageInput';

interface Message {
  role: string;
  content: string;
}

const formatMessage = (message: Message): Message => {
  const lines = message.content.split('\n');
  const lastLine = () => lines.length > 0 ? lines[lines.length - 1] : undefined;
  const firstLine = () => lines.length > 0 ? lines[0] : undefined;

  if (lastLine()?.trim().startsWith('? for shortcuts')) {
    lines.pop();
  }
  if (lastLine()?.trim().includes('───────────────')) {
    lines.pop();
  }
  if (lastLine()?.trim().includes('>')) {
    lines.pop();
  }
  if (lastLine()?.trim().includes('───────────────')) {
    lines.pop();
  }
  if (lastLine()?.trim() === "") {
    lines.pop();
  }
  if (firstLine()?.includes('>')) {
    lines.shift();
  }
  if (firstLine()?.trim() === "") {
    lines.shift();
  }
  return {
    role: message.role,
    content: lines.join('\n'),
  };
}

export default function ChatInterface() {
  const [rawMessages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [serverStatus, setServerStatus] = useState<string>('unknown');
  
  const messages = rawMessages.map(formatMessage);
  
  // Set up polling for messages and server status
  useEffect(() => {
    // Check server status initially
    checkServerStatus();
    
    // Set up polling intervals
    const messageInterval = setInterval(fetchMessages, 1000);
    const statusInterval = setInterval(checkServerStatus, 250);
    
    // Clean up intervals on component unmount
    return () => {
      clearInterval(messageInterval);
      clearInterval(statusInterval);
    };
  }, []);
  
  // Fetch messages from server
  const fetchMessages = async () => {
    try {
      const response = await fetch('http://localhost:8080/messages');
      const data = await response.json();
      if (data.messages) {
        setMessages(data.messages);
      }
    } catch (error) {
      console.error('Error fetching messages:', error);
    }
  };
  
  // Check server status
  const checkServerStatus = async () => {
    try {
      const response = await fetch('http://localhost:8080/status');
      const data = await response.json();
      setServerStatus(data.status);
    } catch (error) {
      console.error('Error checking server status:', error);
      setServerStatus('offline');
    }
  };
  
  // Send a new message
  const sendMessage = async (content: string, type: 'user' | 'raw' = 'user') => {
    // For user messages, require non-empty content
    if (type === 'user' && !content.trim()) return;
    
    // For raw messages, don't set loading state as it's usually fast
    if (type === 'user') {
      setLoading(true);
    }
    
    try {
      const response = await fetch('http://localhost:8080/message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          content, 
          type 
        }),
      });
      
      if (response.ok) {
        // If successful, fetch the updated messages
        // For raw messages, we wait a bit longer to ensure terminal has processed the command
        if (type === 'raw') {
          setTimeout(fetchMessages, 100);
        } else {
          fetchMessages();
        }
      } else {
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
          <span className={`w-2 h-2 rounded-full mr-2 ${'bg-green-500'}`}></span>
          Status: {serverStatus}
        </span>
      </div>
      
      <MessageList messages={messages} />
      
      <MessageInput onSendMessage={sendMessage} disabled={loading} />
    </div>
  );
}