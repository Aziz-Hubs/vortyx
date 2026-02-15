"use client";

import { useMemo, useState } from "react";
import { createClient } from "@connectrpc/connect";
import { VortyxService } from "@/gen/proto/ts/vortyx/v1/service_connect";
import { transport } from "@/lib/transport";

export default function Home() {
  const [inputValue, setInputValue] = useState("");
  const [response, setResponse] = useState("");

  const client = useMemo(() => createClient(VortyxService, transport), []);

  const handleClick = async () => {
    try {
      const res = await client.ping({ message: inputValue });
      setResponse(res.message);
    } catch (e) {
      console.error(e);
      setResponse("Error: " + String(e));
    }
  };

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-8 row-start-2 items-center sm:items-start">
        <h1 className="text-2xl font-bold">Vortyx Ping Test</h1>
        <div className="flex gap-2">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Enter message"
            className="border p-2 rounded text-black"
          />
          <button 
            onClick={handleClick} 
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded transition-colors"
          >
            Ping
          </button>
        </div>
        {response && (
          <div className="p-4 bg-gray-100 rounded text-black w-full">
            <p className="font-semibold">Response:</p>
            <p>{response}</p>
          </div>
        )}
      </main>
    </div>
  );
}
