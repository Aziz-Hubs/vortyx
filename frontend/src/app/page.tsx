import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRight, Shield, Globe, Zap } from "lucide-react";

export default function LandingPage() {
  return (
    <div className="flex flex-col min-h-screen bg-black text-white selection:bg-purple-500/30">
      <header className="px-4 lg:px-6 h-16 flex items-center border-b border-white/10 backdrop-blur-md fixed w-full z-50 bg-black/50">
        <Link className="flex items-center justify-center font-bold text-xl tracking-tighter" href="#">
          <span className="bg-gradient-to-r from-purple-400 to-pink-600 bg-clip-text text-transparent">Vortyx</span>
        </Link>
        <nav className="ml-auto flex gap-4 sm:gap-6">
          <Link className="text-sm font-medium hover:text-purple-400 transition-colors" href="#features">Features</Link>
          <Link className="text-sm font-medium hover:text-purple-400 transition-colors" href="#security">Security</Link>
          <Link className="text-sm font-medium hover:text-purple-400 transition-colors" href="/login">Login</Link>
        </nav>
      </header>
      <main className="flex-1 pt-16">
        <section className="w-full py-24 md:py-32 lg:py-48 xl:py-64 relative overflow-hidden">
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-purple-900/20 via-black to-black"></div>
            <div className="container px-4 md:px-6 relative z-10">
                <div className="flex flex-col items-center space-y-4 text-center">
                    <div className="space-y-2">
                        <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl lg:text-7xl/none bg-gradient-to-br from-white to-gray-400 bg-clip-text text-transparent">
                          Unified Monolith for <br/> MSPs & MSSPs
                        </h1>
                        <p className="mx-auto max-w-[700px] text-gray-400 md:text-xl lg:text-2xl mt-4">
                          Complete visibility, control, and security across your entire infrastructure. One platform, infinite possibilities.
                        </p>
                    </div>
                    <div className="space-x-4 mt-8">
                        <Button asChild size="lg" className="bg-purple-600 hover:bg-purple-700 text-white border-0">
                            <Link href="/login">Get Started <ArrowRight className="ml-2 h-4 w-4" /></Link>
                        </Button>
                        <Button variant="outline" size="lg" asChild className="border-gray-700 text-gray-300 hover:bg-gray-800 hover:text-white">
                            <Link href="#features">Learn More</Link>
                        </Button>
                    </div>
                </div>
            </div>
        </section>

        <section id="features" className="w-full py-12 md:py-24 lg:py-32 bg-zinc-950">
          <div className="container px-4 md:px-6">
            <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
              <div className="flex flex-col items-start space-y-3 p-6 rounded-lg border border-white/5 bg-white/5 hover:bg-white/10 transition-colors">
                <Globe className="h-8 w-8 text-purple-500" />
                <h3 className="text-xl font-bold">Global Visibility</h3>
                <p className="text-gray-400">Real-time telemetry and monitoring across all your endpoints and networks.</p>
              </div>
              <div className="flex flex-col items-start space-y-3 p-6 rounded-lg border border-white/5 bg-white/5 hover:bg-white/10 transition-colors">
                <Zap className="h-8 w-8 text-pink-500" />
                <h3 className="text-xl font-bold">Instant Control</h3>
                <p className="text-gray-400">Execute commands, scripts, and updates instantly via our high-performance RPC.</p>
              </div>
              <div className="flex flex-col items-start space-y-3 p-6 rounded-lg border border-white/5 bg-white/5 hover:bg-white/10 transition-colors">
                <Shield className="h-8 w-8 text-blue-500" />
                <h3 className="text-xl font-bold">Advanced Security</h3>
                <p className="text-gray-400">Integrated SIEM, EDR, and vulnerability scanning to keep your assets safe.</p>
              </div>
            </div>
          </div>
        </section>
      </main>
      <footer className="flex flex-col gap-2 sm:flex-row py-6 w-full shrink-0 items-center px-4 md:px-6 border-t border-white/10">
        <p className="text-xs text-gray-500">© 2026 Vortyx Inc. All rights reserved.</p>
        <nav className="sm:ml-auto flex gap-4 sm:gap-6">
          <Link className="text-xs hover:underline underline-offset-4 text-gray-500" href="#">Terms of Service</Link>
          <Link className="text-xs hover:underline underline-offset-4 text-gray-500" href="#">Privacy</Link>
        </nav>
      </footer>
    </div>
  )
}
