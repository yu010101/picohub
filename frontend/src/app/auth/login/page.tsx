"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(email, password);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-sm px-4 pt-20">
      <h1 className="text-xl font-bold">Login</h1>

      <form onSubmit={handleSubmit} className="mt-6 space-y-4">
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="w-full border border-sand-300 bg-transparent px-3 py-2 text-sm focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Password
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            className="w-full border border-sand-300 bg-transparent px-3 py-2 text-sm focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
          />
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full border border-sand-900 py-2 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 disabled:opacity-40 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
        >
          {loading ? "..." : "Login"}
        </button>
      </form>

      <p className="mt-4 text-sm text-sand-400">
        No account?{" "}
        <Link href="/auth/register" className="text-sand-700 underline dark:text-sand-300">
          Register
        </Link>
      </p>

      <div className="mt-8 border-t border-sand-200 pt-4 font-mono text-xs text-sand-400 dark:border-sand-800 dark:text-sand-600">
        <div>admin@picohub.dev / admin123</div>
        <div>tanaka@example.com / password123</div>
      </div>
    </div>
  );
}
