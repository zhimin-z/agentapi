"use client";

import { useState, useEffect, useRef } from "react";
import MessageList from "./MessageList";
import MessageInput from "./MessageInput";
import { useSearchParams } from "next/navigation";

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
  // null port gets converted to NaN
  const parsedPort = parseInt(searchParams.get("port") as string);
  // We're setting port via URL query param, not directly with setPort
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [port, setPort] = useState<number>(
    isNaN(parsedPort) ? 3284 : parsedPort
  );
  const [portInput, setPortInput] = useState<string>(port.toString());
  const AgentAPIUrl = `http://localhost:${port}`;
  const eventSourceRef = useRef<EventSource | null>(null);

  // Update portInput when port changes
  useEffect(() => {
    setPortInput(port.toString());
  }, [port]);

  // Set up SSE connection to the events endpoint
  useEffect(() => {
    // Function to create and set up EventSource
    const setupEventSource = () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }

      // Reset messages when establishing a new connection
      setMessages([]);

      const eventSource = new EventSource(`${AgentAPIUrl}/events`);
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
      eventSource.close();
    };
  }, [AgentAPIUrl]);

  const [error, setError] = useState<string | null>(null);

  // Send a new message
  const sendMessage = async (
    content: string,
    type: "user" | "raw" = "user"
  ) => {
    // For user messages, require non-empty content
    if (type === "user" && !content.trim()) return;

    // Clear any previous errors
    setError(null);

    // For raw messages, don't set loading state as it's usually fast
    if (type === "user") {
      setLoading(true);
    }

    try {
      const response = await fetch(`${AgentAPIUrl}/message`, {
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
        setError(`Failed to send message: ${fullDetail}`);
        // Auto-clear error after 5 seconds
        setTimeout(() => setError(null), 5000);
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

      setError(`Error sending message: ${fullDetail}`);
      // Auto-clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    } finally {
      if (type === "user") {
        setLoading(false);
      }
    }
  };

  const updatePort = () => {
    const newPort = parseInt(portInput);
    if (!isNaN(newPort) && newPort > 0 && newPort < 65536) {
      window.location.href = `?port=${newPort}`;
    } else {
      setError("Invalid port number. Please enter a number between 1-65535.");
      setTimeout(() => setError(null), 5000);
    }
  };

  return (
    <div className="flex flex-col h-[80vh] bg-gray-100 rounded-lg overflow-hidden border border-gray-300 shadow-lg w-full max-w-[95vw]">
      <div className="p-3 bg-gray-800 text-white text-sm flex items-center justify-between">
        <span>AgentAPI Chat</span>
        <div className="flex items-center space-x-3">
          <div className="flex items-center">
            <label htmlFor="port-input" className="text-white mr-1">
              Port:
            </label>
            <input
              id="port-input"
              type="text"
              value={portInput}
              onChange={(e) => setPortInput(e.target.value)}
              className="w-16 px-1 py-0.5 text-xs rounded border border-gray-400 bg-gray-700 text-white"
              onKeyDown={(e) => e.key === "Enter" && updatePort()}
            />
            <button
              onClick={updatePort}
              className="ml-1 px-2 py-0.5 text-xs bg-gray-600 hover:bg-gray-500 rounded"
            >
              Apply
            </button>
          </div>
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
      </div>

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
              API server is offline. Please start the API server on localhost:
              {port}.
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

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-2 text-sm relative">
          <span className="block sm:inline">{error}</span>
          <button
            onClick={() => setError(null)}
            className="absolute top-0 bottom-0 right-0 px-4 py-2"
          >
            ×
          </button>
        </div>
      )}

      <MessageList messages={messages} loading={loading} />

      <MessageInput onSendMessage={sendMessage} disabled={loading} />
    </div>
  );
}
