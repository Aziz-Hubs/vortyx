export default function DashboardPage() {
  return (
    <div className="prose dark:prose-invert max-w-none">
      <h1>Welcome to Vortyx</h1>
      <p className="lead">
        Select an application from the sidebar to get started.
      </p>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-8">
        <div className="p-6 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">MSP Services</h2>
          <p className="text-gray-600 dark:text-gray-400">Manage your clients, assets, and infrastructure.</p>
        </div>
        <div className="p-6 bg-white dark:bg-gray-800 rounded-lg shadow border border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold mb-2">MSSP Services</h2>
          <p className="text-gray-600 dark:text-gray-400">Monitor threats, compliance, and security posture.</p>
        </div>
      </div>
    </div>
  );
}
