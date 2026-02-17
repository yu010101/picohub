import { ReactNode } from "react";

export function Badge({
  children,
  variant = "default",
}: {
  children: ReactNode;
  variant?: "default" | "green" | "red";
}) {
  const styles = {
    default: "text-sand-500 dark:text-sand-400",
    green: "text-mint-600 dark:text-mint-400",
    red: "text-red-600 dark:text-red-400",
  };

  return (
    <span className={`font-mono text-xs ${styles[variant]}`}>
      [{children}]
    </span>
  );
}
