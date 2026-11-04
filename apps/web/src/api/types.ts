// Типы, зеркально повторяющие DTO бэкенда (services/api/.../dto/dto.go).

export interface User {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string | null;
  region: string | null; // NOTE: в миграции region — text, а не region_id
  bio: string | null;
  joined_at: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  parent_id: string | null;
}

export type TopicType = 'discussion' | 'question' | 'report';

export interface Topic {
  id: string;
  category_id: string;
  author_id: string;
  title: string;
  slug: string;
  body: string;
  type: TopicType;
  pinned: boolean;
  locked: boolean;
  created_at: string;
  last_activity_at: string;
  replies_count: number;
  views_count: number;
  author_name: string;
  category_name: string;
}

export interface Post {
  id: string;
  topic_id: string;
  author_id: string;
  body: string;
  created_at: string;
  edited_at: string | null;
  parent_post_id: string | null;
  author_name: string;
}

export interface Report {
  id: string;
  topic_id: string;
  author_id: string;
  title: string;
  body: string;
  water_id: string | null;
  fished_at: string; // YYYY-MM-DD
  location_visibility: 'water_only' | 'private';
  species: string[];
  media_urls: string[];
  created_at: string;
  author_name: string;
}

export type WaterKind = 'river' | 'lake' | 'reservoir' | 'pond' | 'other';

export interface Water {
  id: string;
  name: string;
  slug: string;
  kind: WaterKind;
  region_id: string | null;
  description: string | null;
  cover_url: string | null;
}

export interface FeedItem {
  kind: 'topic' | 'report';
  topic: Topic;
  report: Report | null;
}

export interface Notification {
  id: string;
  type: string;
  title: string;
  created_at: string;
  read: boolean;
  target_url: string;
}

export interface Page<T> {
  items: T[];
  next_cursor: string | null;
}

export type TopicPage = Page<Topic>;
export type PostPage = Page<Post>;
export type ReportPage = Page<Report>;
export type WaterPage = Page<Water>;
export type FeedPage = Page<FeedItem>;
export type NotificationPage = Page<Notification>;

export interface SearchResults {
  topics: Topic[];
  reports: Report[];
  waters: Water[];
  users: User[];
}

export interface ApiErrorBody {
  error: { code: string; message: string; request_id: string };
}
