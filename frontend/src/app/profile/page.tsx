"use client";

import { useAuth } from "@/hooks/useAuth";
import Link from "next/link";

export default function ProfilePage() {
  const { user, loading, logout } = useAuth();

  if (loading) {
    return (
      <p className="py-20 text-center font-mono text-sm text-sand-400">
        loading...
      </p>
    );
  }

  if (!user) {
    return (
      <div className="mx-auto max-w-lg px-4 pt-20 text-center">
        <p className="text-sand-500">
          <Link href="/auth/login" className="text-sand-900 underline dark:text-sand-100">
            Login
          </Link>{" "}
          to view your profile.
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-8">
      <h1 className="text-xl font-bold">{user.display_name || user.username}</h1>
      <p className="mt-1 font-mono text-sm text-sand-400">@{user.username}</p>

      <dl className="mt-6 space-y-3 border-t border-sand-200 pt-6 text-sm dark:border-sand-800">
        <div className="flex justify-between">
          <dt className="text-sand-500 dark:text-sand-400">Email</dt>
          <dd className="font-mono">{user.email}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-sand-500 dark:text-sand-400">Bio</dt>
          <dd className="text-right">{user.bio || "-"}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-sand-500 dark:text-sand-400">Role</dt>
          <dd className="font-mono">{user.is_admin ? "admin" : "user"}</dd>
        </div>
      </dl>

      <div className="mt-8 border-t border-sand-200 pt-6 dark:border-sand-800">
        <button
          onClick={logout}
          className="text-sm text-red-600 hover:underline"
        >
          Logout
        </button>
      </div>
    </div>
  );
}
