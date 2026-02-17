"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import { Review, Skill } from "@/types";
import { StarRating } from "@/components/StarRating";
import { Badge } from "@/components/Badge";
import { useAuth } from "@/hooks/useAuth";

export default function SkillDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const { user } = useAuth();
  const [skill, setSkill] = useState<Skill | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);

  const [rating, setRating] = useState(5);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [reviewError, setReviewError] = useState("");

  useEffect(() => {
    Promise.all([api.getSkill(slug), api.getReviews(slug)])
      .then(([s, r]) => {
        setSkill(s);
        setReviews(r);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [slug]);

  const handleDownload = () => {
    window.open(api.getDownloadUrl(slug), "_blank");
    if (skill) {
      setSkill({ ...skill, download_count: skill.download_count + 1 });
    }
  };

  const handleSubmitReview = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setReviewError("");
    try {
      const review = await api.createReview(slug, rating, title, body);
      setReviews([review, ...reviews]);
      setTitle("");
      setBody("");
      setRating(5);
    } catch (err) {
      setReviewError(
        err instanceof Error ? err.message : "Failed to submit review"
      );
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <p className="py-20 text-center font-mono text-sm text-sand-400">
        loading...
      </p>
    );
  }

  if (!skill) {
    return (
      <p className="py-20 text-center font-mono text-sm text-sand-400">
        skill not found
      </p>
    );
  }

  const tags: string[] = (() => {
    try {
      return JSON.parse(skill.tags);
    } catch {
      return [];
    }
  })();

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <div className="grid gap-10 lg:grid-cols-[1fr_280px]">
        {/* Main */}
        <div className="min-w-0">
          <div className="flex items-baseline gap-3">
            <h1 className="text-2xl font-bold">{skill.name}</h1>
            <span className="font-mono text-sm text-sand-400">
              v{skill.version}
            </span>
          </div>

          <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-sand-500 dark:text-sand-400">
            <span>{skill.author_name}</span>
            <span className="text-sand-300 dark:text-sand-700">/</span>
            <span className="font-mono">{skill.category}</span>
            <span className="text-sand-300 dark:text-sand-700">/</span>
            <div className="flex items-center gap-1">
              <StarRating rating={skill.avg_rating} size="sm" />
              <span>
                {skill.avg_rating.toFixed(1)} ({skill.review_count})
              </span>
            </div>
            <span className="text-sand-300 dark:text-sand-700">/</span>
            <span>{skill.download_count.toLocaleString()} downloads</span>
          </div>

          {tags.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-2">
              {tags.map((tag) => (
                <span
                  key={tag}
                  className="font-mono text-xs text-sand-400 before:content-['#'] dark:text-sand-500"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Description */}
          <div className="mt-8 border-t border-sand-200 pt-6 dark:border-sand-800">
            <p className="leading-relaxed text-sand-700 dark:text-sand-300">
              {skill.description}
            </p>
          </div>

          {/* Reviews */}
          <div className="mt-10 border-t border-sand-200 pt-6 dark:border-sand-800">
            <h2 className="text-sm font-medium uppercase tracking-wider text-sand-400 dark:text-sand-500">
              Reviews ({reviews.length})
            </h2>

            {user && (
              <form
                onSubmit={handleSubmitReview}
                className="mt-4 border border-sand-200 p-4 dark:border-sand-800"
              >
                <div className="mb-3">
                  <StarRating
                    rating={rating}
                    interactive
                    onChange={setRating}
                  />
                </div>
                <input
                  type="text"
                  placeholder="Title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="mb-2 w-full border border-sand-300 bg-transparent px-3 py-2 text-sm placeholder:text-sand-400 focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
                />
                <textarea
                  placeholder="Comment"
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  rows={3}
                  className="mb-2 w-full border border-sand-300 bg-transparent px-3 py-2 text-sm placeholder:text-sand-400 focus:border-sand-900 focus:outline-none dark:border-sand-700 dark:focus:border-sand-300"
                />
                {reviewError && (
                  <p className="mb-2 text-sm text-red-600">{reviewError}</p>
                )}
                <button
                  type="submit"
                  disabled={submitting || !title || !body}
                  className="border border-sand-900 px-3 py-1.5 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 disabled:opacity-40 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
                >
                  {submitting ? "Posting..." : "Post review"}
                </button>
              </form>
            )}

            <div className="mt-4 divide-y divide-sand-200 dark:divide-sand-800">
              {reviews.length === 0 ? (
                <p className="py-6 text-sm text-sand-400">No reviews yet.</p>
              ) : (
                reviews.map((review) => (
                  <div key={review.id} className="py-4">
                    <div className="flex items-center gap-3 text-sm">
                      <span className="font-medium">{review.username}</span>
                      <StarRating rating={review.rating} size="sm" />
                      <span className="font-mono text-xs text-sand-400">
                        {new Date(review.created_at).toLocaleDateString(
                          "ja-JP"
                        )}
                      </span>
                    </div>
                    <h4 className="mt-1 text-sm font-medium">
                      {review.title}
                    </h4>
                    <p className="mt-1 text-sm text-sand-500 dark:text-sand-400">
                      {review.body}
                    </p>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <aside className="lg:pt-0">
          <div className="border border-sand-200 p-5 dark:border-sand-800 lg:sticky lg:top-20">
            <button
              onClick={handleDownload}
              className="w-full border border-sand-900 py-2.5 text-sm font-medium hover:bg-sand-900 hover:text-sand-50 dark:border-sand-200 dark:hover:bg-sand-200 dark:hover:text-sand-900"
            >
              Download
            </button>

            <dl className="mt-5 space-y-3 text-sm">
              {[
                ["Version", skill.version],
                ["Category", skill.category],
                ["Author", skill.author_name],
                ["Downloads", skill.download_count.toLocaleString()],
                [
                  "Published",
                  new Date(skill.created_at).toLocaleDateString("ja-JP"),
                ],
              ].map(([label, value]) => (
                <div key={label} className="flex justify-between">
                  <dt className="text-sand-400 dark:text-sand-500">{label}</dt>
                  <dd className="font-mono">{value}</dd>
                </div>
              ))}
              <div className="flex justify-between">
                <dt className="text-sand-400 dark:text-sand-500">Scan</dt>
                <dd>
                  <Badge
                    variant={skill.scan_status === "clean" ? "green" : "red"}
                  >
                    {skill.scan_status}
                  </Badge>
                </dd>
              </div>
            </dl>

            <div className="mt-5 border-t border-sand-200 pt-4 dark:border-sand-800">
              <div className="font-mono text-xs text-sand-400 dark:text-sand-500">
                Install
              </div>
              <div className="mt-1.5 border border-sand-200 bg-sand-100 px-3 py-2 font-mono text-xs dark:border-sand-800 dark:bg-sand-900">
                <span className="select-none text-sand-400 dark:text-sand-600">
                  ${" "}
                </span>
                picoclaw install {skill.slug}
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
