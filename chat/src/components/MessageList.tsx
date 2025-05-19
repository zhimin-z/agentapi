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
  const LoadingDots = () => (
    <div className="flex space-x-1">
      <div
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:0ms]`}
      />
      <div
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:300ms]`}
      />
      <div
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:600ms]`}
      />
    </div>
  );

  return (
    <div className="flex-1 overflow-y-auto py-4 flex flex-col gap-4">
      {messages.map((message) => (
        <div
          key={message.id}
          className={`${message.role === "user" ? "text-right" : ""}`}
        >
          <div
            className={`inline-block rounded-lg ${
              message.role === "user"
                ? "bg-accent-foreground rounded-lg max-w-[90%] p-4 text-accent"
                : "max-w-[90%]"
            }`}
          >
            <div
              className={`whitespace-pre-wrap break-words text-left text-sm ${
                message.role === "user" ? "" : "font-mono"
              }`}
            >
              {message.role !== "user" && message.content === "" ? (
                <LoadingDots />
              ) : (
                message.content
              )}
            </div>
          </div>
        </div>
      ))}

      {/* Loading indicator for message being sent */}
      {loading && (
        <div className="w-fit self-end">
          <LoadingDots />
        </div>
      )}

      <div ref={messagesEndRef} />
    </div>
  );
}
