import { Suspense } from "react";
import ChatInterface from "@/components/ChatInterface";

export default function Home() {
  return (
    <Suspense
      fallback={
        <div className="text-center p-4">Loading chat interface...</div>
      }
    >
      <ChatInterface />
    </Suspense>
  );
}
