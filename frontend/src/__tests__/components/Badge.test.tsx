import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Badge } from "@/components/Badge";

describe("Badge", () => {
  it("renders children in brackets", () => {
    render(<Badge>clean</Badge>);
    expect(screen.getByText(/clean/)).toBeInTheDocument();
  });

  it("applies green variant class", () => {
    const { container } = render(<Badge variant="green">ok</Badge>);
    const span = container.querySelector("span");
    expect(span?.className).toContain("text-mint-600");
  });

  it("applies red variant class", () => {
    const { container } = render(<Badge variant="red">flagged</Badge>);
    const span = container.querySelector("span");
    expect(span?.className).toContain("text-red-600");
  });

  it("applies default variant when none specified", () => {
    const { container } = render(<Badge>test</Badge>);
    const span = container.querySelector("span");
    expect(span?.className).toContain("text-sand-500");
  });
});
