"use client";

import { useMemo, useState, useEffect } from "react";
import { createClient } from "@connectrpc/connect";
// @ts-ignore
import { PulseService } from "@/gen/proto/ts/vortyx/pulse/v1/service_connect";
import { transport } from "@/lib/transport";

export default function PulsePage() {
  const [status, setStatus] = useState<string>("Loading...");
  const client = useMemo(() => createClient(PulseService, transport), []);

  useEffect(() => {
    // @ts-ignore
    client.getStatus({}).then((res) => {
      setStatus(res.status);
    }).catch((err: any) => {
      setStatus("Error: " + err.message);
    });
  }, [client]);

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold mb-4">Remote Monitoring & Management (Pulse)</h1>
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <p className="text-lg">System Status: <span className="font-mono font-bold text-green-500">{status}</span></p>
      </div>
    </div>
  );
}
