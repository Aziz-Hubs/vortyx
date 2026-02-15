import Link from 'next/link';
import { auth, signOut } from "@/auth";
import { Button } from "@/components/ui/button";

const apps = [
  { name: 'pulse', label: 'Pulse (RMM)' },
  { name: 'pilot', label: 'Pilot (PSA)' },
  { name: 'nexus', label: 'Nexus (Docs)' },
  { name: 'horizon', label: 'Horizon (Strategy)' },
  { name: 'control', label: 'Control (SaaS)' },
  { name: 'optic', label: 'Optic (CCTV)' },
  { name: 'grid', label: 'Grid (Network)' },
  { name: 'radar', label: 'Radar (SIEM)' },
  { name: 'guard', label: 'Guard (EDR)' },
  { name: 'shield', label: 'Shield (GRC)' },
  { name: 'mind', label: 'Mind (Training)' },
  { name: 'probe', label: 'Probe (Scanner)' },
  { name: 'reflex', label: 'Reflex (SOAR)' },
  { name: 'sonar', label: 'Sonar (NDR)' },
  { name: 'signal', label: 'Signal (Intel)' },
];

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await auth();

  return (
    <div className="flex h-screen bg-gray-100 dark:bg-gray-900">
      {/* Sidebar */}
      <aside className="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <Link href="/" className="text-2xl font-bold text-purple-600">Vortyx</Link>
        </div>
        <nav className="flex-1 overflow-y-auto p-4 space-y-1">
          <Link href="/" className="block px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded hover:bg-gray-100 dark:hover:bg-gray-700">
            Home
          </Link>

          <div className="pt-4 pb-2">
            <p className="px-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Admin</p>
          </div>
          <Link
            href="/admin/users"
            className="block px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            User Management
          </Link>
          <Link
            href="/admin/roles"
            className="block px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            Role Management
          </Link>
          <Link
            href="/admin/audit"
            className="block px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            Audit Logs
          </Link>

          <div className="pt-4 pb-2">
            <p className="px-4 text-xs font-semibold text-gray-500 uppercase tracking-wider">Apps</p>
          </div>
          {apps.map((app) => (
            <Link
              key={app.name}
              href={`/${app.name}`}
              className="block px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              {app.label}
            </Link>
          ))}
        </nav>
        <div className="p-4 border-t border-gray-200 dark:border-gray-700">
          <div className="text-xs text-gray-500 mb-2">
            Logged in as <span className="font-medium text-gray-900 dark:text-gray-100">{session?.user?.name || session?.user?.email || "Guest"}</span>
          </div>
          <form
            action={async () => {
              "use server"
              await signOut()
            }}
          >
            <Button variant="outline" size="sm" className="w-full">Sign Out</Button>
          </form>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto p-8">
        {children}
      </main>
    </div>
  );
}
