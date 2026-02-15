import { signIn } from "@/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Shield } from "lucide-react";

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-black text-white p-4">
      <Card className="w-full max-w-md border-zinc-800 bg-zinc-950 text-white">
        <CardHeader className="space-y-1">
          <div className="flex justify-center mb-4">
            <div className="p-3 rounded-full bg-purple-500/10">
                <Shield className="h-8 w-8 text-purple-500" />
            </div>
          </div>
          <CardTitle className="text-2xl text-center font-bold">Welcome back</CardTitle>
          <CardDescription className="text-center text-zinc-400">
            Sign in to your Vortyx account
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <form
            action={async () => {
              "use server"
              await signIn("zitadel", { redirectTo: "/" })
            }}
          >
            <Button className="w-full bg-white text-black hover:bg-gray-200" type="submit">
              Sign in with SSO (Zitadel)
            </Button>
          </form>
          
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-zinc-800" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-zinc-950 px-2 text-zinc-500">Or continue with</span>
            </div>
          </div>

          <form
            action={async (formData) => {
              "use server"
              await signIn("credentials", formData)
            }}
            className="grid gap-4"
          >
            <div className="grid gap-2">
              <Label htmlFor="email">Username / Email</Label>
              <Input 
                id="username" 
                name="username"
                type="text" 
                placeholder="m@example.com" 
                className="bg-zinc-900 border-zinc-800" 
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input 
                id="password" 
                name="password"
                type="password" 
                className="bg-zinc-900 border-zinc-800" 
                required
              />
            </div>
            <Button className="w-full bg-purple-600 hover:bg-purple-700" type="submit">
              Sign In
            </Button>
          </form>
        </CardContent>
        <CardFooter>
          <p className="text-xs text-center text-gray-500 w-full">
            Note: Username/Password requires Password Grant enabled in Zitadel.
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}
