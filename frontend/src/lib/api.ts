import { AuthResponse, PaginatedResponse, Review, Skill, User } from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
  private getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("token");
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    if (!(options.body instanceof FormData)) {
      headers["Content-Type"] = "application/json";
    }

    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: "Request failed" }));
      throw new Error(error.error || "Request failed");
    }

    return res.json();
  }

  // Auth
  async register(
    username: string,
    email: string,
    password: string,
    displayName?: string
  ): Promise<AuthResponse> {
    return this.request("/auth/register", {
      method: "POST",
      body: JSON.stringify({
        username,
        email,
        password,
        display_name: displayName,
      }),
    });
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    return this.request("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  async me(): Promise<User> {
    return this.request("/auth/me");
  }

  // Skills
  async listSkills(params?: {
    q?: string;
    category?: string;
    sort?: string;
    page?: number;
    per_page?: number;
  }): Promise<PaginatedResponse<Skill>> {
    const searchParams = new URLSearchParams();
    if (params?.q) searchParams.set("q", params.q);
    if (params?.category) searchParams.set("category", params.category);
    if (params?.sort) searchParams.set("sort", params.sort);
    if (params?.page) searchParams.set("page", String(params.page));
    if (params?.per_page) searchParams.set("per_page", String(params.per_page));
    const qs = searchParams.toString();
    return this.request(`/skills${qs ? `?${qs}` : ""}`);
  }

  async getFeaturedSkills(): Promise<Skill[]> {
    return this.request("/skills/featured");
  }

  async getSkill(slug: string): Promise<Skill> {
    return this.request(`/skills/${slug}`);
  }

  async getCategories(): Promise<string[]> {
    return this.request("/skills/categories");
  }

  async uploadSkill(file: File): Promise<Skill> {
    const formData = new FormData();
    formData.append("file", file);
    return this.request("/skills", {
      method: "POST",
      body: formData,
    });
  }

  async updateSkill(
    slug: string,
    data: { description?: string; is_featured?: boolean }
  ): Promise<Skill> {
    return this.request(`/skills/${slug}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteSkill(slug: string): Promise<void> {
    return this.request(`/skills/${slug}`, { method: "DELETE" });
  }

  getDownloadUrl(slug: string): string {
    return `${API_BASE}/skills/${slug}/download`;
  }

  // Reviews
  async getReviews(slug: string): Promise<Review[]> {
    return this.request(`/skills/${slug}/reviews`);
  }

  async createReview(
    slug: string,
    rating: number,
    title: string,
    body: string
  ): Promise<Review> {
    return this.request(`/skills/${slug}/reviews`, {
      method: "POST",
      body: JSON.stringify({ rating, title, body }),
    });
  }
}

export const api = new ApiClient();
