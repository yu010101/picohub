import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { SkillCard } from "@/components/SkillCard";
import { Skill } from "@/types";

const mockSkill: Skill = {
  id: 1,
  slug: "test-skill",
  name: "Test Skill",
  description: "A test skill for unit testing",
  version: "1.0.0",
  category: "testing",
  author_id: 1,
  author_name: "testuser",
  file_hash: "abc123",
  scan_status: "clean",
  download_count: 42,
  is_featured: false,
  tags: '["unit-test","vitest"]',
  avg_rating: 4.5,
  review_count: 3,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
};

describe("SkillCard", () => {
  it("renders skill name", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("Test Skill")).toBeInTheDocument();
  });

  it("renders skill description", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("A test skill for unit testing")).toBeInTheDocument();
  });

  it("renders version", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("v1.0.0")).toBeInTheDocument();
  });

  it("renders category", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("testing")).toBeInTheDocument();
  });

  it("renders author name", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("testuser")).toBeInTheDocument();
  });

  it("renders download count", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("42 dl")).toBeInTheDocument();
  });

  it("renders tags as hashtags", () => {
    render(<SkillCard skill={mockSkill} />);
    expect(screen.getByText("unit-test")).toBeInTheDocument();
    expect(screen.getByText("vitest")).toBeInTheDocument();
  });

  it("links to skill detail page", () => {
    render(<SkillCard skill={mockSkill} />);
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/skills/test-skill");
  });

  it("handles invalid tags JSON gracefully", () => {
    const skillWithBadTags = { ...mockSkill, tags: "not-json" };
    render(<SkillCard skill={skillWithBadTags} />);
    expect(screen.getByText("Test Skill")).toBeInTheDocument();
  });
});
