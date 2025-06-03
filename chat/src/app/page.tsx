import { Chat } from "@/components/chat";
import { ChatProvider } from "@/components/chat-provider";
import { Header } from "./header";

export default function Home() {
  return (
    <ChatProvider>
      <div className="flex flex-col h-svh">
        <Header />
        <Chat />
      </div>
    </ChatProvider>
  );
}
