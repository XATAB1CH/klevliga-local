import { act, cleanup, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { api } from './client';
import { useInfiniteFeed, useInfiniteNotifications, useInfinitePosts, useInfiniteReports, useInfiniteTopics, useInfiniteWaters, useLogin, useLogout, useReply } from './hooks';
import type { NotificationPage, Post, Topic, User } from './types';

const clients: QueryClient[] = [];
function setup(user: User | null = null) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 30_000 }, mutations: { retry: false } } });
  clients.push(client);
  client.setQueryData(['me'], user);
  const wrapper = ({ children }: { children: ReactNode }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  return { client, wrapper };
}
function makeUser(id: string): User {
  return { id, username: id, display_name: id, avatar_url: null, region: null, bio: null, joined_at: '2026-01-01T00:00:00Z' };
}
function makeTopic(id: string): Topic {
  return { id, category_id: 'category', author_id: 'author', title: id, slug: id, body: id, type: 'discussion', pinned: false, locked: false,
    created_at: '2026-01-01T00:00:00Z', last_activity_at: '2026-01-01T00:00:00Z', replies_count: 0, views_count: 0, author_name: 'Рыболов', category_name: 'Обсуждения' };
}
function makePost(index: number): Post {
  return { id: String(index), topic_id: 'topic', author_id: 'author', body: `Ответ ${index}`, author_name: 'Рыболов',
    created_at: '2026-01-01T00:00:00Z', edited_at: null, parent_post_id: null };
}
function notificationPage(title: string, next: string | null = null): NotificationPage {
  return { items: [{ id: title, title, type: 'reply', read: false, target_url: '', created_at: '2026-01-01T00:00:00Z' }], next_cursor: next };
}
afterEach(() => {
  cleanup();
  clients.forEach((client) => client.clear());
  clients.length = 0;
  vi.restoreAllMocks();
});

describe('cursor pagination', () => {
  it('appends feed pages using the returned opaque cursor and stops at the end', async () => {
    const feed = vi.spyOn(api, 'getFeed')
      .mockResolvedValueOnce({ items: [{ kind: 'topic', topic: makeTopic('one'), report: null }], next_cursor: 'opaque-next' })
      .mockResolvedValueOnce({ items: [{ kind: 'topic', topic: makeTopic('two'), report: null }], next_cursor: null });
    const { wrapper } = setup();
    const { result } = renderHook(() => useInfiniteFeed({ kind: 'topic', limit: 1 }), { wrapper });
    await waitFor(() => expect(result.current.data?.pages[0].items[0].topic.id).toBe('one'));
    expect(result.current.hasNextPage).toBe(true);
    await act(async () => { await result.current.fetchNextPage(); });
    await waitFor(() => expect(result.current.data?.pages).toHaveLength(2));
    expect(result.current.data?.pages.flatMap((page) => page.items.map((item) => item.topic.id))).toEqual(['one', 'two']);
    expect(feed).toHaveBeenNthCalledWith(2, { kind: 'topic', limit: 1, cursor: 'opaque-next' }, expect.objectContaining({ aborted: false }));
    expect(result.current.hasNextPage).toBe(false);
    await act(async () => { await result.current.fetchNextPage(); });
    expect(feed).toHaveBeenCalledTimes(2);
  });

  it('preserves category and water filters when loading later topic and report pages', async () => {
    const topics = vi.spyOn(api, 'getTopics')
      .mockResolvedValueOnce({ items: [], next_cursor: 'topics-next' })
      .mockResolvedValueOnce({ items: [], next_cursor: null });
    const reports = vi.spyOn(api, 'getReports')
      .mockResolvedValueOnce({ items: [], next_cursor: 'reports-next' })
      .mockResolvedValueOnce({ items: [], next_cursor: null });
    const { wrapper } = setup();
    const { result } = renderHook(() => ({ topics: useInfiniteTopics({ category_id: 'category-a', limit: 2 }),
      reports: useInfiniteReports({ water_id: 'water-a', limit: 3 }) }), { wrapper });
    await waitFor(() => {
      expect(result.current.topics.hasNextPage).toBe(true);
      expect(result.current.reports.hasNextPage).toBe(true);
    });
    await act(async () => { await Promise.all([result.current.topics.fetchNextPage(), result.current.reports.fetchNextPage()]); });
    expect(topics).toHaveBeenNthCalledWith(2, { category_id: 'category-a', limit: 2, cursor: 'topics-next' }, expect.objectContaining({ aborted: false }));
    expect(reports).toHaveBeenNthCalledWith(2, { water_id: 'water-a', limit: 3, cursor: 'reports-next' }, expect.objectContaining({ aborted: false }));
  });

  it('keeps loaded waters after a later page fails and retries that page with the same filters', async () => {
    const waters = vi.spyOn(api, 'getWaters')
      .mockResolvedValueOnce({ items: [], next_cursor: 'waters-next' })
      .mockRejectedValueOnce(new Error('Сеть недоступна'))
      .mockResolvedValueOnce({ items: [], next_cursor: null });
    const { wrapper } = setup();
    const { result } = renderHook(() => useInfiniteWaters({ q: 'озеро', region_id: 'region-a', limit: 2 }), { wrapper });
    await waitFor(() => expect(result.current.data?.pages).toHaveLength(1));
    expect(result.current.isFetchNextPageError).toBe(false);
    await act(async () => { await result.current.fetchNextPage(); });
    await waitFor(() => expect(result.current.isFetchNextPageError).toBe(true));
    expect(result.current.data?.pages).toHaveLength(1);
    await act(async () => { await result.current.fetchNextPage(); });
    await waitFor(() => expect(result.current.data?.pages).toHaveLength(2));
    expect(waters).toHaveBeenNthCalledWith(3, { q: 'озеро', region_id: 'region-a', limit: 2, cursor: 'waters-next' }, expect.objectContaining({ aborted: false }));
    expect(result.current.hasNextPage).toBe(false);
  });

  it('refreshes every loaded page after publishing in a topic with more than twenty replies', async () => {
    const stored = Array.from({ length: 21 }, (_, index) => makePost(index + 1));
    vi.spyOn(api, 'getPosts').mockImplementation(async (_topicId, params = {}) => {
      const start = Number(params.cursor ?? 0);
      const limit = params.limit ?? 20;
      return { items: stored.slice(start, start + limit), next_cursor: start + limit < stored.length ? String(start + limit) : null };
    });
    vi.spyOn(api, 'reply').mockImplementation(async () => {
      const post = makePost(stored.length + 1);
      stored.push(post);
      return post;
    });
    const { wrapper } = setup();
    const { result } = renderHook(() => ({ posts: useInfinitePosts('topic'), reply: useReply('topic') }), { wrapper });
    await waitFor(() => expect(result.current.posts.data?.pages[0].items).toHaveLength(20));
    await act(async () => { await result.current.posts.fetchNextPage(); });
    await waitFor(() => expect(result.current.posts.data?.pages.flatMap((page) => page.items)).toHaveLength(21));
    await act(async () => { const published = await result.current.reply.mutateAsync({ body: 'Ответ 22' }); expect(published.id).toBe('22'); });
    await waitFor(() => expect(result.current.posts.data?.pages.flatMap((page) => page.items)).toHaveLength(22));
    const ids = result.current.posts.data!.pages.flatMap((page) => page.items.map((post) => post.id));
    expect(new Set(ids).size).toBe(22);
    expect(ids.at(-1)).toBe('22');
  });

  it('aborts a personal page request on logout and starts the next account from its first page', async () => {
    let previousSignal: AbortSignal | undefined;
    const notifications = vi.spyOn(api, 'getNotifications')
      .mockResolvedValueOnce(notificationPage('alice-first', 'alice-next'))
      .mockImplementationOnce((_params, signal) => new Promise((_resolve, reject) => {
        previousSignal = signal;
        signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')), { once: true });
      }))
      .mockResolvedValueOnce(notificationPage('bob-first'));
    vi.spyOn(api, 'logout').mockResolvedValue(undefined);
    vi.spyOn(api, 'login').mockResolvedValue(undefined);
    vi.spyOn(api, 'getMe').mockResolvedValue(makeUser('bob'));
    const { wrapper } = setup(makeUser('alice'));
    const { result } = renderHook(() => ({ notifications: useInfiniteNotifications(), logout: useLogout(), login: useLogin() }), { wrapper });
    await waitFor(() => expect(result.current.notifications.data?.pages[0].items[0].title).toBe('alice-first'));
    let nextPage!: Promise<unknown>;
    act(() => { nextPage = result.current.notifications.fetchNextPage(); });
    await waitFor(() => expect(previousSignal).toBeDefined());
    await act(async () => { await result.current.logout.mutateAsync(); await nextPage; });
    expect(previousSignal?.aborted).toBe(true);
    await act(async () => { await result.current.login.mutateAsync({ email: 'bob@example.com', password: 'a-valid-password' }); });
    await waitFor(() => expect(result.current.notifications.data?.pages[0].items[0].title).toBe('bob-first'));
    expect(result.current.notifications.data?.pages).toHaveLength(1);
    expect(notifications).toHaveBeenNthCalledWith(3, { cursor: undefined }, expect.objectContaining({ aborted: false }));
  });

  it('does not fetch personal pages before authentication', () => {
    const notifications = vi.spyOn(api, 'getNotifications');
    const { wrapper } = setup();
    const { result } = renderHook(() => useInfiniteNotifications(), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(notifications).not.toHaveBeenCalled();
  });
});
