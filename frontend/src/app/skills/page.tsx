"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { Skill } from "@/types";
import { SkillCard } from "@/components/SkillCard";
import { useDebounce } from "@/hooks/useDebounce";

export default function SkillsPage() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("");
  const [sort, setSort] = useState("newest");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);

  const debouncedQuery = useDebounce(query, 300);

  const fetchSkills = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.listSkills({
        q: debouncedQuery || undefined,
        category: category || undefined,
        sort,
        page,
        per_page: 12,
      });
      setSkills(res.data || []);
      setTotalPages(res.total_pages);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [debouncedQuery, category, sort, page]);

  useEffect(() => {
    fetchSkills();
  }, [fetchSkills]);

  useEffect(() => {
    api.getCategories().then(setCategories).catch(console.error);
  }, []);

  useEffect(() => {
    setPage(1);
  }, [debouncedQuery, category, sort]);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="text-2xl font-bold">Skills</h1>

      {/* Filters */}
      <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
        <input
          type="text"
          placeholder="Search..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="border border-sand-300 bg-transparent px-3 py-2 text-sm placeholder:text-sand-400 focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:placeholder:text-sand-600 dark:focus:border-sand-300 sm:w-64"
        />
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          className="border border-sand-300 bg-transparent px-3 py-2 text-sm focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
        >
          <option value="">All categories</option>
          {categories.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value)}
          className="border border-sand-300 bg-transparent px-3 py-2 text-sm focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
        >
          <option value="newest">Newest</option>
          <option value="downloads">Downloads</option>
          <option value="rating">Rating</option>
          <option value="name">Name</option>
        </select>
      </div>

      {/* Results */}
      <div className="mt-8">
        {loading ? (
          <p className="py-12 text-center font-mono text-sm text-sand-400">
            loading...
          </p>
        ) : skills.length > 0 ? (
          <>
            <div className="divide-y divide-sand-200 border-y border-sand-200 dark:divide-sand-800 dark:border-sand-800">
              {skills.map((skill) => (
                <SkillCard key={skill.id} skill={skill} />
              ))}
            </div>

            {totalPages > 1 && (
              <div className="mt-6 flex items-center justify-between font-mono text-sm">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="text-sand-500 hover:text-sand-900 disabled:text-sand-300 dark:disabled:text-sand-700"
                >
                  &larr; prev
                </button>
                <span className="text-sand-400">
                  {page}/{totalPages}
                </span>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="text-sand-500 hover:text-sand-900 disabled:text-sand-300 dark:disabled:text-sand-700"
                >
                  next &rarr;
                </button>
              </div>
            )}
          </>
        ) : (
          <p className="py-12 text-center font-mono text-sm text-sand-400">
            no skills found
          </p>
        )}
      </div>
    </div>
  );
}
