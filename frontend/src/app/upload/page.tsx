"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import Link from "next/link";

export default function UploadPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");
  const [dragActive, setDragActive] = useState(false);

  if (authLoading) return null;

  if (!user) {
    return (
      <div className="mx-auto max-w-lg px-4 pt-20 text-center">
        <p className="text-sand-500">
          <Link href="/auth/login" className="text-sand-900 underline dark:text-sand-100">
            Login
          </Link>{" "}
          to publish skills.
        </p>
      </div>
    );
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;
    setUploading(true);
    setError("");
    try {
      const skill = await api.uploadSkill(file);
      router.push(`/skills/${skill.slug}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setUploading(false);
    }
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(e.type === "dragenter" || e.type === "dragover");
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (e.dataTransfer.files?.[0]) setFile(e.dataTransfer.files[0]);
  };

  return (
    <div className="mx-auto max-w-lg px-4 py-8">
      <h1 className="text-xl font-bold">Publish skill</h1>

      <form onSubmit={handleSubmit} className="mt-6">
        <div
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
          className={`border p-8 text-center transition-colors ${
            dragActive
              ? "border-mint-500 bg-mint-500/5"
              : "border-sand-300 dark:border-sand-700"
          }`}
        >
          {file ? (
            <div className="font-mono text-sm">
              <div>{file.name}</div>
              <div className="mt-1 text-xs text-sand-400">
                {(file.size / 1024 / 1024).toFixed(2)} MB
              </div>
              <button
                type="button"
                onClick={() => setFile(null)}
                className="mt-2 text-xs text-red-600 hover:underline"
              >
                remove
              </button>
            </div>
          ) : (
            <div>
              <p className="text-sm text-sand-500 dark:text-sand-400">
                Drop .zip here or{" "}
                <label className="cursor-pointer text-sand-900 underline dark:text-sand-100">
                  browse
                  <input
                    type="file"
                    accept=".zip"
                    className="hidden"
                    onChange={(e) =>
                      e.target.files?.[0] && setFile(e.target.files[0])
                    }
                  />
                </label>
              </p>
              <p className="mt-1 font-mono text-xs text-sand-400 dark:text-sand-600">
                zip, max 10MB, must contain manifest.json
              </p>
            </div>
          )}
        </div>

        {error && (
          <p className="mt-3 text-sm text-red-600">{error}</p>
        )}

        <button
          type="submit"
          disabled={!file || uploading}
          className="mt-4 w-full border border-sand-900 py-2.5 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 disabled:opacity-40 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
        >
          {uploading ? "Uploading..." : "Publish"}
        </button>
      </form>

      <div className="mt-8 border-t border-sand-200 pt-6 dark:border-sand-800">
        <div className="font-mono text-xs text-sand-400 dark:text-sand-600">
          <div className="mb-2 text-sand-500 dark:text-sand-400">
            Required package structure:
          </div>
          <div className="border border-sand-200 bg-sand-100 px-3 py-2 leading-relaxed dark:border-sand-800 dark:bg-sand-900">
            <div>
              <span className="text-sand-500">your-skill/</span>
            </div>
            <div>
              {"  "}
              <span className="text-mint-600 dark:text-mint-400">
                manifest.json
              </span>{" "}
              &larr; required
            </div>
            <div>{"  "}SKILL.md</div>
            <div>{"  "}main.py</div>
          </div>
        </div>
      </div>
    </div>
  );
}
