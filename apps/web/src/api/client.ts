// Типизированный HTTP-клиент к Go API. Все запросы шлют сессионную cookie.
import type {
  Category, FeedPage, NotificationPage, Post, PostPage, Report, ReportPage,
  SearchResults, Topic, TopicPage, User, Water, WaterPage, SessionUser, ManagedUser, Role, AuditEvent, AuditFilters, Page,
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

async function apiFetch<T>(path: string, init?: RequestInit, expectJson = true): Promise<T> {
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
  if (!expectJson) return undefined as T;
  if (body === null) {
    throw new ApiError(res.status, 'INVALID_RESPONSE', 'Сервер вернул непонятный ответ. Попробуйте ещё раз.');
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
  getMe: (signal?: AbortSignal) => apiFetch<SessionUser>('/auth/me', { signal }),
  login: (email: string, password: string) => apiFetch<void>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  register: (email: string, username: string, password: string) => apiFetch<void>('/auth/register', { method: 'POST', body: JSON.stringify({ email, username, password }) }, false),
  verifyRegistration: (email: string, code: string) => apiFetch<void>('/auth/register/verify', { method: 'POST', body: JSON.stringify({ email, code }) }),
  logout: () => apiFetch<void>('/auth/logout', { method: 'POST' }),
  getUser: (id: string, signal?: AbortSignal) => apiFetch<User>(`/users/${id}`, { signal }),
  uploadAvatar: (file: File) => apiFetch<User>('/auth/avatar', {
    method: 'PUT', body: file, headers: { 'Content-Type': file.type },
  }),

  // Reads
  getFeed: (p: FeedParams = {}, signal?: AbortSignal) => apiFetch<FeedPage>(`/feed${qs({ cursor: p.cursor, limit: p.limit, kind: p.kind === 'all' ? undefined : p.kind })}`, { signal }),
  getCategories: (signal?: AbortSignal) => apiFetch<Category[]>('/forum/categories', { signal }),
  getTopics: (p: TopicsParams = {}, signal?: AbortSignal) => apiFetch<TopicPage>(`/forum/topics${qs({ cursor: p.cursor, limit: p.limit, category_id: p.category_id })}`, { signal }),
  getTopic: (id: string, signal?: AbortSignal) => apiFetch<Topic>(`/forum/topics/${id}`, { signal }),
  getPosts: (topicId: string, p: { cursor?: string; limit?: number } = {}, signal?: AbortSignal) => apiFetch<PostPage>(`/forum/topics/${topicId}/posts${qs({ cursor: p.cursor, limit: p.limit })}`, { signal }),
  getReports: (p: ReportsParams = {}, signal?: AbortSignal) => apiFetch<ReportPage>(`/reports${qs({ cursor: p.cursor, limit: p.limit, water_id: p.water_id })}`, { signal }),
  getReport: (id: string, signal?: AbortSignal) => apiFetch<Report>(`/reports/${id}`, { signal }),
  getWaters: (p: WatersParams = {}, signal?: AbortSignal) => apiFetch<WaterPage>(`/waters${qs({ cursor: p.cursor, limit: p.limit, q: p.q, region_id: p.region_id })}`, { signal }),
  getWater: (id: string, signal?: AbortSignal) => apiFetch<Water>(`/waters/${id}`, { signal }),
  search: (p: SearchParams, signal?: AbortSignal) => apiFetch<SearchResults>(`/search${qs({ q: p.q, type: p.type, limit: p.limit })}`, { signal }),
  getNotifications: (p: { cursor?: string; limit?: number } = {}, signal?: AbortSignal) => apiFetch<NotificationPage>(`/notifications${qs({ cursor: p.cursor, limit: p.limit })}`, { signal }),

  // Writes
  searchManagedUsers: (q: string, signal?: AbortSignal) => apiFetch<{ items: ManagedUser[] }>(`/admin/users${qs({ q, limit: 20 })}`, { signal }),
  getManagedUser: (id: string, signal?: AbortSignal) => apiFetch<ManagedUser>(`/admin/users/${id}`, { signal }),
  changeRoles: (id: string, roles: Role[], reason: string) => json<void>(`/admin/users/${id}/roles`, 'PUT', { roles, reason }),
  changeUserStatus: (id: string, status: 'active' | 'blocked', reason: string) => json<void>(`/admin/users/${id}/status`, 'PUT', { status, reason }),
  getAuditEvents: (p: AuditFilters & { cursor?: string; limit?: number }, signal?: AbortSignal) => apiFetch<Page<AuditEvent>>(`/admin/audit-events${qs({ ...p })}`, { signal }),
  createTopic: (categoryId: string, title: string, body: string) => json<Topic>('/forum/topics', 'POST', { category_id: categoryId, title, body }),
  reply: (topicId: string, body: string, parentPostId?: string | null) => json<Post>(`/forum/topics/${topicId}/posts`, 'POST', { body, parent_post_id: parentPostId ?? null }),
  createReport: (p: { title: string; body: string; water_id?: string | null; fished_at: string; location_visibility: string; species: string[] }) => json<Report>('/reports', 'POST', p),
  follow: (targetType: string, targetId: string) => apiFetch<void>('/follows', { method: 'POST', body: JSON.stringify({ target_type: targetType, target_id: targetId }) }),
  unfollow: (targetType: string, targetId: string) => apiFetch<void>(`/follows/${targetType}/${targetId}`, { method: 'DELETE' }),
  flag: (targetType: string, targetId: string, reason: string) => apiFetch<void>('/moderation/flags', { method: 'POST', body: JSON.stringify({ target_type: targetType, target_id: targetId, reason }) }, false),
};
