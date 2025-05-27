"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import MessageList from "./MessageList";
import MessageInput from "./MessageInput";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { Button } from "./ui/button";
import { TriangleAlertIcon } from "lucide-react";
import { Alert, AlertTitle, AlertDescription } from "./ui/alert";
import { ModeToggle } from "./mode-toggle";

interface Message {
  id: number;
  role: string;
  content: string;
}

// Draft messages are used to optmistically update the UI
// before the server responds.
interface DraftMessage extends Omit<Message, "id"> {
  id?: number;
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

const isDraftMessage = (message: Message | DraftMessage): boolean => {
  return message.id === undefined;
};

export default function ChatInterface() {
  const [messages, setMessages] = useState<(Message | DraftMessage)[]>([]);
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
          // Clean up draft messages
          const updatedMessages = [...prevMessages].filter(
            (m) => !isDraftMessage(m)
          );

          // Check if message with this ID already exists
          const existingIndex = updatedMessages.findIndex(
            (m) => m.id === data.id
          );

          if (existingIndex !== -1) {
            // Update existing message
            updatedMessages[existingIndex] = {
              role: data.role,
              content: data.message,
              id: data.id,
            };
            return updatedMessages;
          } else {
            // Add new message
            return [
              ...updatedMessages,
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
      setMessages((prevMessages) => [
        ...prevMessages,
        { role: "user", content },
      ]);
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
        setMessages((prevMessages) =>
          prevMessages.filter((m) => !isDraftMessage(m))
        );
        setLoading(false);
      }
    }
  };

  return (
    <div className="flex flex-col h-svh">
      <header className="p-4 flex items-center justify-between border-b">
        <span className="font-bold">AgentAPI Chat</span>

        <div className="flex items-center gap-4">
          {serverStatus !== "unknown" && (
            <div className="flex items-center gap-2 text-sm font-medium">
              <span
                className={`text-secondary w-2 h-2 rounded-full ${
                  ["offline", "unknown"].includes(serverStatus)
                    ? "bg-red-500 ring-2 ring-red-500/35"
                    : "bg-green-500 ring-2 ring-green-500/35"
                }`}
              />
              <span className="sr-only">Status:</span>
              <span className="first-letter:uppercase">{serverStatus}</span>
            </div>
          )}
          <ModeToggle />
        </div>
      </header>

      <main className="flex flex-1 flex-col w-full overflow-auto">
        {serverStatus === "offline" && (
          <div className="p-4 w-full max-w-4xl mx-auto">
            <Alert className="flex border-yellow-500">
              <TriangleAlertIcon className="h-4 w-4  stroke-yellow-600" />
              <div>
                <AlertTitle>API server is offline</AlertTitle>
                <AlertDescription>
                  Please start the AgentAPI server. Attempting to connect to:{" "}
                  {agentAPIUrl || "N/A"}.
                </AlertDescription>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="ml-auto"
                onClick={() => window.location.reload()}
              >
                Retry
              </Button>
            </Alert>
          </div>
        )}

        <MessageList messages={messages} />
        <MessageInput onSendMessage={sendMessage} disabled={loading} />
      </main>
    </div>
  );
}
