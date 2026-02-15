"use client";

import { useMemo, useState, useEffect } from "react";
import { createClient } from "@connectrpc/connect";
// @ts-ignore
import { SonarService } from "@/gen/proto/ts/vortyx/sonar/v1/service_connect";
import { transport } from "@/lib/transport";

export default function SonarPage() {
  const [status, setStatus] = useState<string>("Loading...");
  const client = useMemo(() => createClient(SonarService, transport), []);

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
      <h1 className="text-3xl font-bold mb-4">Network Detection (Sonar)</h1>
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <p className="text-lg">System Status: <span className="font-mono font-bold text-green-500">{status}</span></p>
      </div>
    </div>
  );
}
