import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { StarRating } from "@/components/StarRating";

describe("StarRating", () => {
  it("renders filled stars based on rating", () => {
    const { container } = render(<StarRating rating={3} />);
    const spans = container.querySelectorAll("span");
    // First span has filled stars, inner span has unfilled
    expect(spans[0].textContent).toBe("*****");
  });

  it("renders 5 interactive buttons", () => {
    render(<StarRating rating={3} interactive onChange={() => {}} />);
    const buttons = screen.getAllByRole("button");
    expect(buttons).toHaveLength(5);
  });

  it("calls onChange when interactive star clicked", () => {
    const onChange = vi.fn();
    render(<StarRating rating={2} interactive onChange={onChange} />);
    const buttons = screen.getAllByRole("button");
    fireEvent.click(buttons[3]); // click 4th star
    expect(onChange).toHaveBeenCalledWith(4);
  });

  it("applies sm size class", () => {
    const { container } = render(<StarRating rating={4} size="sm" />);
    const span = container.querySelector("span");
    expect(span?.className).toContain("text-xs");
  });
});
