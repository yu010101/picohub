"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { Skill } from "@/types";
import { SkillCard } from "@/components/SkillCard";

export default function Home() {
  const [featured, setFeatured] = useState<Skill[]>([]);

  useEffect(() => {
    api.getFeaturedSkills().then(setFeatured).catch(console.error);
  }, []);

  return (
    <div className="mx-auto max-w-6xl px-4">
      {/* Hero - terminal style, no gradient */}
      <section className="pb-12 pt-16 sm:pt-24">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          PicoClaw skills,
          <br />
          curated and scanned.
        </h1>
        <p className="mt-4 max-w-lg text-sand-500 dark:text-sand-400">
          $10 RISC-Vボード用AIエージェントのスキルレジストリ。
          アップロードされた全パッケージはマルウェアスキャン済み。
        </p>

        <div className="mt-8 inline-block border border-sand-200 bg-sand-100 px-4 py-3 font-mono text-sm dark:border-sand-800 dark:bg-sand-900">
          <span className="select-none text-sand-400 dark:text-sand-600">
            ${" "}
          </span>
          <span className="text-mint-600 dark:text-mint-400">
            picoclaw install
          </span>{" "}
          line-messenger
        </div>

        <div className="mt-6 flex gap-4">
          <Link
            href="/skills"
            className="border border-sand-900 px-4 py-2 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
          >
            Browse all skills
          </Link>
          <Link
            href="/upload"
            className="px-4 py-2 text-sm text-sand-500 hover:text-sand-900 dark:text-sand-400 dark:hover:text-sand-100"
          >
            Publish yours
          </Link>
        </div>
      </section>

      {/* Featured */}
      {featured.length > 0 && (
        <section className="border-t border-sand-200 py-10 dark:border-sand-800">
          <div className="mb-6 flex items-baseline justify-between">
            <h2 className="text-sm font-medium uppercase tracking-wider text-sand-400 dark:text-sand-500">
              Featured
            </h2>
            <Link
              href="/skills"
              className="font-mono text-xs text-sand-400 hover:text-sand-700 dark:text-sand-500 dark:hover:text-sand-300"
            >
              all &rarr;
            </Link>
          </div>

          <div className="divide-y divide-sand-200 border-y border-sand-200 dark:divide-sand-800 dark:border-sand-800 sm:divide-y-0 sm:border-y-0 sm:grid sm:grid-cols-2 sm:gap-px sm:bg-sand-200 sm:border sm:dark:bg-sand-800">
            {featured.map((skill) => (
              <div
                key={skill.id}
                className="bg-sand-50 dark:bg-sand-950 sm:bg-sand-50 sm:dark:bg-sand-950"
              >
                <SkillCard skill={skill} />
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Quick reference instead of "How it works" */}
      <section className="border-t border-sand-200 py-10 dark:border-sand-800">
        <h2 className="mb-6 text-sm font-medium uppercase tracking-wider text-sand-400 dark:text-sand-500">
          Quick reference
        </h2>
        <div className="grid gap-6 sm:grid-cols-2">
          <div className="font-mono text-sm">
            <div className="mb-2 text-xs text-sand-400 dark:text-sand-500">
              Install a skill
            </div>
            <div className="border border-sand-200 bg-sand-100 px-3 py-2 dark:border-sand-800 dark:bg-sand-900">
              <span className="select-none text-sand-400 dark:text-sand-600">
                ${" "}
              </span>
              picoclaw install{" "}
              <span className="text-mint-600 dark:text-mint-400">
                &lt;slug&gt;
              </span>
            </div>
          </div>
          <div className="font-mono text-sm">
            <div className="mb-2 text-xs text-sand-400 dark:text-sand-500">
              Publish a skill
            </div>
            <div className="border border-sand-200 bg-sand-100 px-3 py-2 dark:border-sand-800 dark:bg-sand-900">
              <span className="select-none text-sand-400 dark:text-sand-600">
                ${" "}
              </span>
              picoclaw publish{" "}
              <span className="text-mint-600 dark:text-mint-400">
                ./my-skill
              </span>
            </div>
          </div>
          <div className="font-mono text-sm">
            <div className="mb-2 text-xs text-sand-400 dark:text-sand-500">
              Package structure
            </div>
            <div className="border border-sand-200 bg-sand-100 px-3 py-2 leading-relaxed dark:border-sand-800 dark:bg-sand-900">
              <div>
                <span className="text-sand-400 dark:text-sand-500">
                  my-skill/
                </span>
              </div>
              <div>
                {"  "}
                <span className="text-mint-600 dark:text-mint-400">
                  manifest.json
                </span>
              </div>
              <div>
                {"  "}SKILL.md
              </div>
              <div>
                {"  "}main.py
              </div>
            </div>
          </div>
          <div className="font-mono text-sm">
            <div className="mb-2 text-xs text-sand-400 dark:text-sand-500">
              Security checks
            </div>
            <div className="border border-sand-200 bg-sand-100 px-3 py-2 leading-relaxed dark:border-sand-800 dark:bg-sand-900">
              <div>
                <span className="text-mint-600 dark:text-mint-400">
                  PASS
                </span>{" "}
                manifest validation
              </div>
              <div>
                <span className="text-mint-600 dark:text-mint-400">
                  PASS
                </span>{" "}
                symlink detection
              </div>
              <div>
                <span className="text-mint-600 dark:text-mint-400">
                  PASS
                </span>{" "}
                malware scan
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
