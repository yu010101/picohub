"use client";

import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";
import { ThemeToggle } from "./ThemeToggle";
import { useState } from "react";

export function Header() {
  const { user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <header className="border-b border-sand-200 dark:border-sand-800">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        <div className="flex items-center gap-8">
          <Link href="/" className="font-mono text-sm font-bold tracking-tight">
            picohub
          </Link>
          <nav className="hidden items-center gap-6 md:flex">
            <Link
              href="/skills"
              className="text-sm text-sand-500 hover:text-sand-900 dark:text-sand-400 dark:hover:text-sand-100"
            >
              Browse
            </Link>
            <Link
              href="/upload"
              className="text-sm text-sand-500 hover:text-sand-900 dark:text-sand-400 dark:hover:text-sand-100"
            >
              Publish
            </Link>
          </nav>
        </div>

        <div className="flex items-center gap-2">
          <ThemeToggle />
          {user ? (
            <div className="relative">
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="font-mono text-sm text-sand-600 hover:text-sand-900 dark:text-sand-400 dark:hover:text-sand-100"
              >
                {user.username}
              </button>
              {menuOpen && (
                <>
                  <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(false)}
                  />
                  <div className="absolute right-0 z-50 mt-2 w-40 border border-sand-200 bg-sand-50 py-1 dark:border-sand-800 dark:bg-sand-950">
                    <Link
                      href="/profile"
                      onClick={() => setMenuOpen(false)}
                      className="block px-3 py-1.5 text-sm text-sand-600 hover:bg-sand-100 dark:text-sand-400 dark:hover:bg-sand-900"
                    >
                      Profile
                    </Link>
                    <button
                      onClick={() => {
                        logout();
                        setMenuOpen(false);
                      }}
                      className="block w-full px-3 py-1.5 text-left text-sm text-sand-600 hover:bg-sand-100 dark:text-sand-400 dark:hover:bg-sand-900"
                    >
                      Logout
                    </button>
                  </div>
                </>
              )}
            </div>
          ) : (
            <Link
              href="/auth/login"
              className="text-sm text-sand-500 hover:text-sand-900 dark:text-sand-400 dark:hover:text-sand-100"
            >
              Login
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
