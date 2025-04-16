"use client";

import { useState, FormEvent, KeyboardEvent, useEffect, useRef } from "react";

interface MessageInputProps {
  onSendMessage: (message: string, type: "user" | "raw") => void;
  disabled?: boolean;
}

interface SentChar {
  char: string;
  id: number;
  timestamp: number;
}

export default function MessageInput({
  onSendMessage,
  disabled = false,
}: MessageInputProps) {
  const [message, setMessage] = useState("");
  const [inputMode, setInputMode] = useState<"text" | "control">("text");
  const [sentChars, setSentChars] = useState<SentChar[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const nextCharId = useRef(0);
  const [controlAreaFocused, setControlAreaFocused] = useState(false);

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSendMessage(message, "user");
      setMessage("");
    }
  };

  // Remove sent characters after they expire (2 seconds)
  useEffect(() => {
    if (sentChars.length === 0) return;

    const interval = setInterval(() => {
      const now = Date.now();
      setSentChars((chars) =>
        chars.filter((char) => now - char.timestamp < 2000)
      );
    }, 100);

    return () => clearInterval(interval);
  }, [sentChars]);

  const addSentChar = (char: string) => {
    const newChar: SentChar = {
      char,
      id: nextCharId.current++,
      timestamp: Date.now(),
    };
    setSentChars((prev) => [...prev, newChar]);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    // In control mode, send special keys as raw messages
    if (inputMode === "control" && !disabled) {
      // List of keys to send as raw input when in control mode
      const specialKeys: Record<string, string> = {
        ArrowUp: "\x1b[A", // Escape sequence for up arrow
        ArrowDown: "\x1b[B", // Escape sequence for down arrow
        ArrowRight: "\x1b[C", // Escape sequence for right arrow
        ArrowLeft: "\x1b[D", // Escape sequence for left arrow
        Escape: "\x1b", // Escape key
        Tab: "\t", // Tab key
        Delete: "\x1b[3~", // Delete key
        Home: "\x1b[H", // Home key
        End: "\x1b[F", // End key
        PageUp: "\x1b[5~", // Page Up
        PageDown: "\x1b[6~", // Page Down
        Backspace: "\b", // Backspace key
      };

      // Check if the pressed key is in our special keys map
      if (specialKeys[e.key]) {
        e.preventDefault();
        addSentChar(e.key);
        onSendMessage(specialKeys[e.key], "raw");
        return;
      }

      // Handle Enter as raw newline when in control mode
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        addSentChar("⏎");
        onSendMessage("\r", "raw");
        return;
      }

      // Handle Ctrl+key combinations
      if (e.ctrlKey) {
        const ctrlMappings: Record<string, string> = {
          c: "\x03", // Ctrl+C (SIGINT)
          d: "\x04", // Ctrl+D (EOF)
          z: "\x1A", // Ctrl+Z (SIGTSTP)
          l: "\x0C", // Ctrl+L (clear screen)
          a: "\x01", // Ctrl+A (beginning of line)
          e: "\x05", // Ctrl+E (end of line)
          w: "\x17", // Ctrl+W (delete word)
          u: "\x15", // Ctrl+U (clear line)
          r: "\x12", // Ctrl+R (reverse history search)
        };

        if (ctrlMappings[e.key.toLowerCase()]) {
          e.preventDefault();
          addSentChar(`Ctrl+${e.key.toUpperCase()}`);
          onSendMessage(ctrlMappings[e.key.toLowerCase()], "raw");
          return;
        }
      }

      // If it's a printable character (length 1), send it as raw input
      if (e.key.length === 1) {
        e.preventDefault();
        addSentChar(e.key);
        onSendMessage(e.key, "raw");
        return;
      }
    } else if (e.key === "Enter" && !e.shiftKey) {
      // Normal Enter handling for text mode with non-empty message
      e.preventDefault();
      handleSubmit(e);
    }
  };

  // Keep focus on the textarea for capturing key events
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.focus();
      if (inputMode === "control") {
        setControlAreaFocused(true);
      }
    }
  }, [inputMode]);

  return (
    <form
      onSubmit={handleSubmit}
      className="border-t border-gray-300 p-4 bg-white"
    >
      <div className="flex flex-col">
        <div className="mb-2 flex items-center">
          <div className="flex rounded-lg overflow-hidden border border-gray-300">
            <button
              type="button"
              onClick={() => setInputMode("text")}
              className={`px-3 py-1 text-sm font-medium ${
                inputMode === "text"
                  ? "bg-blue-500 text-white"
                  : "bg-gray-100 text-gray-700 hover:bg-gray-200"
              }`}
            >
              Text
            </button>
            <button
              type="button"
              onClick={() => setInputMode("control")}
              className={`px-3 py-1 text-sm font-medium ${
                inputMode === "control"
                  ? "bg-blue-500 text-white"
                  : "bg-gray-100 text-gray-700 hover:bg-gray-200"
              }`}
            >
              Control
            </button>
          </div>
          <span className="ml-3 text-xs text-gray-600">
            Switch to <span className="font-medium">Control</span> mode to send
            raw keystrokes (↑,↓,Tab,Ctrl+C,Ctrl+R) directly to the terminal
          </span>
        </div>

        {inputMode === "control" && !disabled && (
          <div className="mb-1 text-xs text-blue-600 font-mono flex justify-between">
            <span>Control mode - keystrokes sent directly to terminal</span>
            {sentChars.length > 0 && (
              <div className="flex space-x-1">
                {sentChars.map((char) => (
                  <span
                    key={char.id}
                    className="font-mono px-1 bg-blue-100 rounded text-blue-800 transition-opacity"
                    style={{
                      opacity: Math.max(
                        0,
                        1 - (Date.now() - char.timestamp) / 2000
                      ),
                    }}
                  >
                    {char.char}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}

        <div className="flex">
          {inputMode === "control" && !disabled ? (
            <div
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              ref={textareaRef as any}
              tabIndex={0}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              onKeyDown={handleKeyDown as any}
              onFocus={() => setControlAreaFocused(true)}
              onBlur={() => setControlAreaFocused(false)}
              className="flex-1 cursor-text border rounded-lg p-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-500 bg-gray-50 border-blue-200 min-h-[3.5rem] flex items-center justify-center"
            >
              {controlAreaFocused
                ? "Press any key to send to terminal (arrows, Ctrl+C, Ctrl+R, etc.)"
                : "Click or focus this area to send keystrokes to terminal"}
            </div>
          ) : (
            <>
              <textarea
                ref={textareaRef}
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={"Type a message..."}
                className="flex-1 resize-none border rounded-l-lg p-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 bg-white"
                rows={2}
                disabled={disabled}
              />
              <button
                type="submit"
                disabled={disabled || !message.trim()}
                className="bg-green-500 text-white px-4 rounded-r-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Send
              </button>
            </>
          )}
        </div>
      </div>
    </form>
  );
}
