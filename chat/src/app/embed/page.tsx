import { Chat } from "@/components/chat";
import { ChatProvider } from "@/components/chat-provider";

export default function EmbedPage() {
  return (
    <ChatProvider>
      <Chat />
    </ChatProvider>
  );
}
