import Link from "next/link";
import { Skill } from "@/types";
import { StarRating } from "./StarRating";

export function SkillCard({ skill }: { skill: Skill }) {
  const tags: string[] = (() => {
    try {
      return JSON.parse(skill.tags);
    } catch {
      return [];
    }
  })();

  return (
    <Link href={`/skills/${skill.slug}`}>
      <article className="group border-b border-sand-200 py-5 last:border-0 dark:border-sand-800 sm:border sm:px-5 sm:py-4 sm:hover:bg-sand-100 sm:dark:hover:bg-sand-900">
        <div className="flex items-baseline justify-between gap-4">
          <h3 className="font-medium text-sand-900 dark:text-sand-100">
            {skill.name}
          </h3>
          <span className="shrink-0 font-mono text-xs text-sand-400 dark:text-sand-600">
            {skill.category}
          </span>
        </div>

        <p className="mt-1.5 line-clamp-2 text-sm text-sand-500 dark:text-sand-400">
          {skill.description}
        </p>

        <div className="mt-3 flex items-center gap-3 text-xs text-sand-400 dark:text-sand-500">
          <span className="font-mono">v{skill.version}</span>
          <span>&middot;</span>
          <span>{skill.author_name}</span>
          <span>&middot;</span>
          <div className="flex items-center gap-1">
            <StarRating rating={skill.avg_rating} size="sm" />
            <span>{skill.avg_rating.toFixed(1)}</span>
          </div>
          <span>&middot;</span>
          <span>{skill.download_count.toLocaleString()} dl</span>
        </div>

        {tags.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {tags.slice(0, 4).map((tag) => (
              <span
                key={tag}
                className="font-mono text-xs text-sand-400 before:content-['#'] dark:text-sand-600"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </article>
    </Link>
  );
}
