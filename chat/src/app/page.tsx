import { Suspense } from 'react';
import ChatInterface from '@/components/ChatInterface';

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-4 md:p-12">
      <div className="w-full max-w-5xl">
        <h1 className="text-3xl font-bold mb-6 text-center">AgentAPI Chat</h1>
        {/* Use Suspense boundary for useSearchParams in ChatInterface */}
        <Suspense fallback={<div className="text-center p-4">Loading chat interface...</div>}>
          <ChatInterface />
        </Suspense>
      </div>
    </main>
  );
}