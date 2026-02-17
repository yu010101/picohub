"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }
    setLoading(true);
    try {
      await register(username, email, password, displayName || undefined);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  };

  const inputClass =
    "w-full border border-sand-300 bg-transparent px-3 py-2 text-sm focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300";

  return (
    <div className="mx-auto max-w-sm px-4 pt-20">
      <h1 className="text-xl font-bold">Register</h1>

      <form onSubmit={handleSubmit} className="mt-6 space-y-4">
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Username
          </label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            className={inputClass}
          />
        </div>
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Display name{" "}
            <span className="text-sand-400 dark:text-sand-600">optional</span>
          </label>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={inputClass}
          />
        </div>
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className={inputClass}
          />
        </div>
        <div>
          <label className="mb-1 block text-sm text-sand-500 dark:text-sand-400">
            Password{" "}
            <span className="text-sand-400 dark:text-sand-600">min 8</span>
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            className={inputClass}
          />
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full border border-sand-900 py-2 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 disabled:opacity-40 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
        >
          {loading ? "..." : "Create account"}
        </button>
      </form>

      <p className="mt-4 text-sm text-sand-400">
        Already have an account?{" "}
        <Link href="/auth/login" className="text-sand-700 underline dark:text-sand-300">
          Login
        </Link>
      </p>
    </div>
  );
}
