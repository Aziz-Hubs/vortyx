import { Suspense } from "react";
import AuthContent from "./auth-content";

export default function AuthPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center bg-black text-white">Loading Auth...</div>}>
      <AuthContent />
    </Suspense>
  );
}
