export interface User {
  id: number;
  username: string;
  email: string;
  display_name: string;
  bio: string;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface Skill {
  id: number;
  slug: string;
  name: string;
  description: string;
  version: string;
  category: string;
  author_id: number;
  author_name: string;
  file_hash: string;
  scan_status: string;
  download_count: number;
  is_featured: boolean;
  tags: string;
  avg_rating: number;
  review_count: number;
  created_at: string;
  updated_at: string;
}

export interface Review {
  id: number;
  skill_id: number;
  user_id: number;
  username: string;
  rating: number;
  title: string;
  body: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface ErrorResponse {
  error: string;
  message?: string;
}
