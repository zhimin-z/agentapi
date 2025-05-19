"use client";

import { useEffect, useRef } from "react";

interface Message {
  role: string;
  content: string;
  id: number;
}

interface MessageListProps {
  messages: Message[];
  loading?: boolean;
}

export default function MessageList({
  messages,
  loading = false,
}: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const prevMessagesLengthRef = useRef<number>(0);
  const prevLoadingRef = useRef<boolean>(false);

  // Enhanced scrolling behavior to handle:
  // 1. New messages being added
  // 2. Loading indicator appearing/disappearing
  // 3. New user messages (to ensure they're always visible)
  useEffect(() => {
    const lastMessage = messages[messages.length - 1];
    const isNewUserMessage =
      messages.length > prevMessagesLengthRef.current &&
      lastMessage?.role === "user";

    const loadingChanged = loading !== prevLoadingRef.current;

    // Store current scroll position and total scroll height
    const messageContainer = messagesEndRef.current?.parentElement;
    if (messagesEndRef.current && messageContainer) {
      // Determine if we should force scroll
      const shouldForceScroll =
        isNewUserMessage || // New user message added
        loading || // Loading indicator is active
        loadingChanged; // Loading state changed

      // Check if we're already near the bottom
      const isNearBottom =
        messageContainer.scrollHeight -
          messageContainer.scrollTop -
          messageContainer.clientHeight <
        100;

      // Scroll if we're forced to or if we're already near the bottom
      if (shouldForceScroll || isNearBottom) {
        messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
      }
    }

    // Update references for next comparison
    prevMessagesLengthRef.current = messages.length;
    prevLoadingRef.current = loading;
  }, [messages, loading]);

  // If no messages, show a placeholder
  if (messages.length === 0) {
    return (
      <div className="flex-1 p-6 flex items-center justify-center text-gray-500 bg-white">
        <p>No messages yet. Start the conversation!</p>
      </div>
    );
  }

  // Define a component for the animated dots
  const LoadingDots = ({ darkMode = false }: { darkMode?: boolean }) => (
    <div className="flex space-x-1">
      <div
        className={`w-2 h-2 rounded-full ${
          darkMode ? "bg-gray-600" : "bg-white"
        } animate-pulse`}
        style={{ animationDelay: "0ms" }}
      ></div>
      <div
        className={`w-2 h-2 rounded-full ${
          darkMode ? "bg-gray-600" : "bg-white"
        } animate-pulse`}
        style={{ animationDelay: "300ms" }}
      ></div>
      <div
        className={`w-2 h-2 rounded-full ${
          darkMode ? "bg-gray-600" : "bg-white"
        } animate-pulse`}
        style={{ animationDelay: "600ms" }}
      ></div>
    </div>
  );

  return (
    <div className="flex-1 overflow-y-auto py-4">
      {messages.map((message) => (
        <div
          key={message.id}
          className={`mb-4 ${message.role === "user" ? "text-right" : ""}`}
        >
          <div
            className={`inline-block  rounded-lg ${
              message.role === "user"
                ? "bg-gray-800 text-white rounded max-w-[90%]"
                : "max-w-[90%]"
            }`}
          >
            <div className="text-xs mb-1 font-bold text-left">
              {message.role === "user" ? "You" : "AgentAPI"}
            </div>
            <div
              className={`whitespace-pre-wrap break-words text-left ${
                message.role === "user" ? "" : "font-mono"
              }`}
            >
              {message.role !== "user" && message.content === "" ? (
                <div className="py-2">
                  <LoadingDots darkMode={true} />
                </div>
              ) : (
                message.content
              )}
            </div>
          </div>
        </div>
      ))}

      {/* Loading indicator for message being sent */}
      {loading && (
        <div className="mb-4 text-right">
          <div className="inline-block px-4 py-2 rounded-lg bg-blue-500 text-white rounded-tr-none">
            <div className="text-xs mb-1 font-bold text-left">You</div>
            <div className="h-6 flex items-center">
              <LoadingDots darkMode={false} />
            </div>
          </div>
        </div>
      )}

      <div ref={messagesEndRef} />
    </div>
  );
}
