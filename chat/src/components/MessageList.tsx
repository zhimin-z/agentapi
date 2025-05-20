"use client";

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
  // If no messages, show a placeholder
  if (messages.length === 0) {
    return (
      <div className="flex-1 p-6 flex items-center justify-center text-muted-foreground">
        <p>No messages yet. Start the conversation!</p>
      </div>
    );
  }

  // Define a component for the animated dots
  const LoadingDots = () => (
    <div className="flex space-x-1">
      <div
        aria-hidden="true"
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:0ms]`}
      />
      <div
        aria-hidden="true"
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:300ms]`}
      />
      <div
        aria-hidden="true"
        className={`size-1 rounded-full bg-foreground animate-pulse [animation-delay:600ms]`}
      />
      <span className="sr-only">Loading...</span>
    </div>
  );

  return (
    <div
      className="overflow-y-auto flex-1"
      ref={(scrollAreaRef) => {
        if (!scrollAreaRef) {
          return;
        }

        scrollAreaRef.scrollTo({
          top: scrollAreaRef.scrollHeight,
        });

        const callback: MutationCallback = (mutationsList) => {
          for (const mutation of mutationsList) {
            if (mutation.type === "childList") {
              scrollAreaRef.scrollTo({
                top: scrollAreaRef.scrollHeight,
                behavior: "smooth",
              });
            }
          }
        };

        const observer = new MutationObserver(callback);
        observer.observe(scrollAreaRef, { childList: true, subtree: false });
      }}
    >
      <div className="p-4 flex flex-col gap-4 max-w-4xl mx-auto">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`${message.role === "user" ? "text-right" : ""}`}
          >
            <div
              className={`inline-block rounded-lg ${
                message.role === "user"
                  ? "bg-accent-foreground rounded-lg max-w-[90%] px-4 py-3 text-accent"
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
      </div>
    </div>
  );
}
