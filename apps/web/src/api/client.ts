// Типизированный HTTP-клиент к Go API. Все запросы шлют сессионную cookie.
import type {
  Category, FeedPage, NotificationPage, Post, PostPage, Report, ReportPage,
  SearchResults, Topic, TopicPage, User, Water, WaterPage,
} from './types';

const BASE = '/api/v1';

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  });

  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);

  if (!res.ok) {
    throw new ApiError(
      res.status,
      body?.error?.code ?? 'UNKNOWN',
      body?.error?.message ?? res.statusText,
    );
  }

  return body as T;
}

type QueryValue = string | number | undefined | null;

function qs(params: Record<string, QueryValue>): string {
  const parts = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`);

  return parts.length ? `?${parts.join('&')}` : '';
}

export interface FeedParams { cursor?: string; limit?: number; kind?: 'all' | 'topic' | 'report' }
export interface TopicsParams { cursor?: string; limit?: number; category_id?: string }
export interface ReportsParams { cursor?: string; limit?: number; water_id?: string }
export interface WatersParams { cursor?: string; limit?: number; q?: string; region_id?: string }
export interface SearchParams { q: string; type?: string; limit?: number }

function json<T>(path: string, method: string, body: unknown): Promise<T> {
  return apiFetch<T>(path, { method, body: JSON.stringify(body) });
}

export const api = {
  // Auth
  getMe: () => apiFetch<User>('/auth/me'),
  login: (email: string, password: string) => apiFetch<void>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  register: (email: string, username: string, password: string) => apiFetch<void>('/auth/register', { method: 'POST', body: JSON.stringify({ email, username, password }) }),
  logout: () => apiFetch<void>('/auth/logout', { method: 'POST' }),
  getUser: (id: string) => apiFetch<User>(`/users/${id}`),

  // Reads
  getFeed: (p: FeedParams = {}) => apiFetch<FeedPage>(`/feed${qs({ cursor: p.cursor, limit: p.limit, kind: p.kind === 'all' ? undefined : p.kind })}`),
  getCategories: () => apiFetch<Category[]>('/forum/categories'),
  getTopics: (p: TopicsParams = {}) => apiFetch<TopicPage>(`/forum/topics${qs({ cursor: p.cursor, limit: p.limit, category_id: p.category_id })}`),
  getTopic: (id: string) => apiFetch<Topic>(`/forum/topics/${id}`),
  getPosts: (topicId: string, p: { cursor?: string; limit?: number } = {}) => apiFetch<PostPage>(`/forum/topics/${topicId}/posts${qs({ cursor: p.cursor, limit: p.limit })}`),
  getReports: (p: ReportsParams = {}) => apiFetch<ReportPage>(`/reports${qs({ cursor: p.cursor, limit: p.limit, water_id: p.water_id })}`),
  getReport: (id: string) => apiFetch<Report>(`/reports/${id}`),
  getWaters: (p: WatersParams = {}) => apiFetch<WaterPage>(`/waters${qs({ cursor: p.cursor, limit: p.limit, q: p.q, region_id: p.region_id })}`),
  getWater: (id: string) => apiFetch<Water>(`/waters/${id}`),
  search: (p: SearchParams) => apiFetch<SearchResults>(`/search${qs({ q: p.q, type: p.type, limit: p.limit })}`),
  getNotifications: (p: { cursor?: string; limit?: number } = {}) => apiFetch<NotificationPage>(`/notifications${qs({ cursor: p.cursor, limit: p.limit })}`),

  // Writes
  createTopic: (categoryId: string, title: string, body: string) => json<Topic>('/forum/topics', 'POST', { category_id: categoryId, title, body }),
  reply: (topicId: string, body: string, parentPostId?: string | null) => json<Post>(`/forum/topics/${topicId}/posts`, 'POST', { body, parent_post_id: parentPostId ?? null }),
  createReport: (p: { title: string; body: string; water_id?: string | null; fished_at: string; location_visibility: string; species: string[] }) => json<Report>('/reports', 'POST', p),
  follow: (targetType: string, targetId: string) => apiFetch<void>('/follows', { method: 'POST', body: JSON.stringify({ target_type: targetType, target_id: targetId }) }),
  unfollow: (targetType: string, targetId: string) => apiFetch<void>(`/follows/${targetType}/${targetId}`, { method: 'DELETE' }),
  flag: (targetType: string, targetId: string, reason: string) => apiFetch<void>('/moderation/flags', { method: 'POST', body: JSON.stringify({ target_type: targetType, target_id: targetId, reason }) }),
};
