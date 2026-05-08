"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function Home() {
  const [roomId, setRoomId] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();

  const createRoom = async () => {
    setLoading(true);
    setError("");
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${apiUrl}/api/rooms`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ttl_minutes: 60 }),
      });

      if (!res.ok) throw new Error("Failed to create room");

      const data = await res.json();
      router.push(`/room/${data.room_id}`);
    } catch (err) {
      setError("Failed to create room. Please try again.");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const joinRoom = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId.trim()) {
      router.push(`/room/${roomId.trim()}`);
    }
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-6 bg-slate-950 text-white">
      <div className="max-w-md w-full space-y-8 text-center">
        <div className="space-y-4">
          <h1 className="text-5xl font-extrabold tracking-tight text-indigo-500">
            GhostRoom
          </h1>
          <p className="text-xl text-slate-400">
            Instant guest video calls. No login, no hassle.
          </p>
        </div>

        <div className="mt-12 space-y-6">
          <button
            onClick={createRoom}
            disabled={loading}
            className="w-full flex items-center justify-center px-8 py-4 border border-transparent text-lg font-medium rounded-xl text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-indigo-500/20"
          >
            {loading ? "Creating..." : "Create Instant Room"}
          </button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-slate-800"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-slate-950 text-slate-500 uppercase tracking-widest">or</span>
            </div>
          </div>

          <form onSubmit={joinRoom} className="space-y-4">
            <div>
              <input
                type="text"
                placeholder="Enter Room ID"
                value={roomId}
                onChange={(e) => setRoomId(e.target.value)}
                className="block w-full px-4 py-4 rounded-xl bg-slate-900 border-slate-800 text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
              />
            </div>
            <button
              type="submit"
              disabled={!roomId.trim()}
              className="w-full flex items-center justify-center px-8 py-4 border border-slate-700 text-lg font-medium rounded-xl text-slate-200 bg-transparent hover:bg-slate-900 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500 transition-all disabled:opacity-50"
            >
              Join Existing Room
            </button>
          </form>
        </div>

        {error && (
          <p className="mt-4 text-red-400 text-sm">{error}</p>
        )}

        <div className="mt-16 grid grid-cols-3 gap-4 text-slate-500 text-sm font-medium">
          <div className="p-4 rounded-lg bg-slate-900/50">
            <div className="mb-1 text-indigo-400">P2P</div>
            WebRTC
          </div>
          <div className="p-4 rounded-lg bg-slate-900/50">
            <div className="mb-1 text-indigo-400">Secure</div>
            Signaling
          </div>
          <div className="p-4 rounded-lg bg-slate-900/50">
            <div className="mb-1 text-indigo-400">Zero</div>
            Onboarding
          </div>
        </div>
      </div>
    </main>
  );
}
