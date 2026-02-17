"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { User } from "@/types";

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      setLoading(false);
      return;
    }
    api
      .me()
      .then(setUser)
      .catch(() => {
        localStorage.removeItem("token");
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.login(email, password);
    localStorage.setItem("token", res.token);
    setUser(res.user);
    return res;
  }, []);

  const register = useCallback(
    async (
      username: string,
      email: string,
      password: string,
      displayName?: string
    ) => {
      const res = await api.register(username, email, password, displayName);
      localStorage.setItem("token", res.token);
      setUser(res.user);
      return res;
    },
    []
  );

  const logout = useCallback(() => {
    localStorage.removeItem("token");
    setUser(null);
  }, []);

  return { user, loading, login, register, logout };
}
