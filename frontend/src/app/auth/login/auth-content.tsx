"use client";

import { useSearchParams, useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Shield, AlertCircle, Loader2 } from "lucide-react";
import { performZitadelLogin, AuthError } from "@/lib/zitadel-session";

export default function AuthContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const authRequestId = searchParams.get("authRequest") || searchParams.get("authRequestID");

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authRequestId) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-black text-white p-4">
        <Card className="w-full max-w-md border-red-900/50 bg-zinc-950 text-white">
          <CardHeader>
            <div className="flex justify-center mb-4">
              <AlertCircle className="h-12 w-12 text-red-500" />
            </div>
            <CardTitle className="text-xl text-center">Invalid Request</CardTitle>
            <CardDescription className="text-center text-zinc-400">
              Missing Auth Request ID. Please try logging in again from the application.
            </CardDescription>
          </CardHeader>
          <CardFooter>
            <Button className="w-full" onClick={() => router.push("/login")}>
              Go to Login
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  const getErrorMessage = (err: AuthError) => {
    const message = err.message || "Authentication failed";
    
    // Map specific Zitadel error messages to user-friendly ones
    if (
      message.includes("Errors.User.InvalidPassword") || 
      message.includes("password check failed") ||
      message.includes("Password is invalid")
    ) {
      return "Incorrect password. Please try again.";
    }
    if (message.includes("Errors.User.NotFound") || message.includes("user not found")) {
      return "User not found. Please check your username.";
    }
    if (message.includes("Errors.User.Locked")) {
      return "Your account has been locked. Please contact support.";
    }
    if (message.includes("Errors.Token.Invalid")) {
      return "Authentication session expired. Please refresh the page.";
    }
    
    return message;
  };

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);
    const formUsername = formData.get("username") as string;
    const formPassword = formData.get("password") as string;

    try {
      const callbackUrl = await performZitadelLogin(formUsername, formPassword, authRequestId);
      window.location.href = callbackUrl;
    } catch (err: any) {
      console.error("Login flow failed:", err);
      setError(getErrorMessage(err));
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-black text-white p-4">
      <Card className="w-full max-w-md border-zinc-800 bg-zinc-950 text-white shadow-2xl">
        <CardHeader className="space-y-1">
          <div className="flex justify-center mb-4">
            <div className="p-3 rounded-full bg-purple-500/10">
              <Shield className="h-10 w-10 text-purple-500" />
            </div>
          </div>
          <CardTitle className="text-2xl text-center font-bold">Vortyx Secure Login</CardTitle>
          <CardDescription className="text-center text-zinc-400">
            Enterprise Identity Gateway
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin} className="grid gap-4">
            {error && (
              <div className="p-3 rounded-md bg-red-500/10 border border-red-500/50 flex items-center gap-2 text-sm text-red-400">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <p>{error}</p>
              </div>
            )}
            <div className="grid gap-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                name="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
                className="bg-zinc-900 border-zinc-800 focus:ring-purple-500"
                required
                disabled={loading}
              />
            </div>
            <div className="grid gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <a href="#" className="text-xs text-purple-400 hover:text-purple-300 transition-colors">
                  Forgot password?
                </a>
              </div>
              <Input
                id="password"
                name="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="bg-zinc-900 border-zinc-800 focus:ring-purple-500"
                required
                disabled={loading}
              />
            </div>
            <Button 
              className="w-full bg-purple-600 hover:bg-purple-700 mt-2 transition-all" 
              type="submit"
              disabled={loading}
            >
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Authenticating...
                </>
              ) : (
                "Sign In"
              )}
            </Button>
          </form>
        </CardContent>
        <CardFooter className="flex flex-col gap-4">
          <div className="relative w-full">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-zinc-800" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-zinc-950 px-2 text-zinc-500 font-medium">Security Verified</span>
            </div>
          </div>
          <p className="text-[10px] text-center text-zinc-600 uppercase tracking-widest">
            Protected by Vortyx Unified Security Mesh
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}
