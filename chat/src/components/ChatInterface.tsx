"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import MessageList from "./MessageList";
import MessageInput from "./MessageInput";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";

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
  const [serverStatus, setServerStatus] = useState<string>("unknown");
  const searchParams = useSearchParams();

  const getAgentApiUrl = useCallback(() => {
    const apiUrlFromParam = searchParams.get("url");
    if (apiUrlFromParam) {
      try {
        // Validate if it's a proper URL
        new URL(apiUrlFromParam);
        return apiUrlFromParam;
      } catch (e) {
        console.warn("Invalid url parameter, defaulting...", e);
        // Fallback if parsing fails or it's not a valid URL.
        // Ensure window is defined (for SSR/Node.js environments during build)
        return typeof window !== "undefined" ? window.location.origin : "";
      }
    }
    // Ensure window is defined
    return typeof window !== "undefined" ? window.location.origin : "";
  }, [searchParams]);

  const [agentAPIUrl, setAgentAPIUrl] = useState<string>(getAgentApiUrl());

  const eventSourceRef = useRef<EventSource | null>(null);

  // Update agentAPIUrl when searchParams change (e.g. url is added/removed)
  useEffect(() => {
    setAgentAPIUrl(getAgentApiUrl());
  }, [getAgentApiUrl, searchParams]);

  // Set up SSE connection to the events endpoint
  useEffect(() => {
    // Function to create and set up EventSource
    const setupEventSource = () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }

      // Reset messages when establishing a new connection
      setMessages([]);

      if (!agentAPIUrl) {
        console.warn(
          "agentAPIUrl is not set, SSE connection cannot be established."
        );
        setServerStatus("offline"); // Or some other appropriate status
        return null; // Don't try to connect if URL is empty
      }

      const eventSource = new EventSource(`${agentAPIUrl}/events`);
      eventSourceRef.current = eventSource;

      // Handle message updates
      eventSource.addEventListener("message_update", (event) => {
        const data: MessageUpdateEvent = JSON.parse(event.data);

        setMessages((prevMessages) => {
          // Check if message with this ID already exists
          const existingIndex = prevMessages.findIndex((m) => m.id === data.id);

          if (existingIndex !== -1) {
            // Update existing message
            const updatedMessages = [...prevMessages];
            updatedMessages[existingIndex] = {
              role: data.role,
              content: data.message,
              id: data.id,
            };
            return updatedMessages;
          } else {
            // Add new message
            return [
              ...prevMessages,
              {
                role: data.role,
                content: data.message,
                id: data.id,
              },
            ];
          }
        });
      });

      // Handle status changes
      eventSource.addEventListener("status_change", (event) => {
        const data: StatusChangeEvent = JSON.parse(event.data);
        setServerStatus(data.status);
      });

      // Handle connection open (server is online)
      eventSource.onopen = () => {
        // Connection is established, but we'll wait for status_change event
        // for the actual server status
        console.log("EventSource connection established - messages reset");
      };

      // Handle connection errors
      eventSource.onerror = (error) => {
        console.error("EventSource error:", error);
        setServerStatus("offline");

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
      if (eventSource) {
        // Check if eventSource was successfully created
        eventSource.close();
      }
    };
  }, [agentAPIUrl]);

  // Send a new message
  const sendMessage = async (
    content: string,
    type: "user" | "raw" = "user"
  ) => {
    // For user messages, require non-empty content
    if (type === "user" && !content.trim()) return;

    // For raw messages, don't set loading state as it's usually fast
    if (type === "user") {
      setLoading(true);
    }

    try {
      const response = await fetch(`${agentAPIUrl}/message`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          content: content,
          type,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        console.error("Failed to send message:", errorData);
        const detail = errorData.detail;
        const messages =
          "errors" in errorData
            ? // eslint-disable-next-line @typescript-eslint/no-explicit-any
              errorData.errors.map((e: any) => e.message).join(", ")
            : "";

        const fullDetail = `${detail}: ${messages}`;
        toast.error(`Failed to send message`, {
          description: fullDetail,
        });
      }
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      console.error("Error sending message:", error);
      const detail = error.detail;
      const messages =
        "errors" in error
          ? // eslint-disable-next-line @typescript-eslint/no-explicit-any
            error.errors.map((e: any) => e.message).join("\n")
          : "";

      const fullDetail = `${detail}: ${messages}`;

      toast.error(`Error sending message`, {
        description: fullDetail,
      });
    } finally {
      if (type === "user") {
        setLoading(false);
      }
    }
  };

  return (
    <div className="flex flex-col h-svh">
      <header className="p-3 dark:text-white text-sm flex items-center justify-between border-b">
        <span className="font-medium">AgentAPI Chat</span>

        <div className="flex items-center space-x-3">
          <div className="flex items-center">
            <span
              className={`w-2 h-2 rounded-full mr-2 ${
                ["offline", "unknown"].includes(serverStatus)
                  ? "bg-red-500"
                  : "bg-green-500"
              }`}
            ></span>
            <span>Status: {serverStatus}</span>
          </div>
        </div>
      </header>

      <main className="flex flex-1 flex-col w-full max-w-4xl mx-auto overflow-auto pb-4 px-2">
        {(serverStatus === "offline" || serverStatus === "unknown") && (
          <div className="bg-yellow-100 border-y border-yellow-400 text-yellow-800 px-4 py-3 flex items-center justify-between font-medium">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 mr-2"
                fill="currentColor"
                viewBox="0 0 20 20"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
              <span>
                API server is offline. Please start the AgentAPI server.
                Attempting to connect to: {agentAPIUrl || "N/A"}.
              </span>
            </div>
            <button
              onClick={() => window.location.reload()}
              className="bg-yellow-200 px-3 py-1 rounded text-xs hover:bg-yellow-300"
            >
              Retry Connection
            </button>
          </div>
        )}

        <MessageList messages={messages} loading={loading} />
        <MessageInput onSendMessage={sendMessage} disabled={loading} />
      </main>
    </div>
  );
}
