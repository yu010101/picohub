interface StarRatingProps {
  rating: number;
  size?: "sm" | "md";
  interactive?: boolean;
  onChange?: (rating: number) => void;
}

export function StarRating({
  rating,
  size = "md",
  interactive = false,
  onChange,
}: StarRatingProps) {
  if (!interactive) {
    const filled = Math.round(rating);
    return (
      <span className={`font-mono ${size === "sm" ? "text-xs" : "text-sm"} text-mint-500`}>
        {"*".repeat(filled)}
        <span className="text-sand-300 dark:text-sand-700">
          {"*".repeat(5 - filled)}
        </span>
      </span>
    );
  }

  return (
    <div className="flex gap-0.5">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          onClick={() => onChange?.(star)}
          className={`font-mono text-lg ${
            star <= rating
              ? "text-mint-500"
              : "text-sand-300 dark:text-sand-700"
          } hover:text-mint-400`}
        >
          *
        </button>
      ))}
    </div>
  );
}
