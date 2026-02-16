"use client";

import { useState, useEffect, useMemo } from "react";
import { createClient } from "@connectrpc/connect";
import { PlatformService } from "@/gen/proto/ts/vortyx/platform/v1/service_connect";
import { transport } from "@/lib/transport";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function AuditPage() {
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const client = useMemo(() => createClient(PlatformService, transport), []);

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const res = await client.getAuditLogs({ page: 1, pageSize: 20 });
        setLogs(res.logs);
      } catch (error) {
        console.error("Failed to fetch logs", error);
      } finally {
        setLoading(false);
      }
    };
    fetchLogs();
  }, [client]);

  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Audit Logs</h1>
        <p className="text-muted-foreground">System activity and security events.</p>
      </div>

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Resource</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center h-24">Loading...</TableCell>
              </TableRow>
            ) : logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center h-24">No logs found.</TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell>{log.createdAt?.toDate().toLocaleString()}</TableCell>
                  <TableCell>{log.username} <span className="text-xs text-gray-500">({log.userId})</span></TableCell>
                  <TableCell>{log.action}</TableCell>
                  <TableCell>{log.resourceType}:{log.resourceId}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
