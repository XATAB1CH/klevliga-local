import { act, cleanup, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from './client';
import { useFeed, useLogin, useLogout, useMe, useNotifications, useRegister, useReports, useSearch, useTopics, useWaters } from './hooks';
import type { NotificationPage, User, Water } from './types';

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

function makeWater(id: string): Water {
  return { id, name: id, slug: id, kind: 'lake', region_id: null, description: null, cover_url: null };
}

function notificationPage(title: string): NotificationPage {
  return { items: [{ id: title, title, type: 'reply', read: false, target_url: '', created_at: '2026-01-01T00:00:00Z' }], next_cursor: null };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => { resolve = done; });
  return { promise, resolve };
}

beforeEach(() => {
  vi.spyOn(api, 'getMe').mockRejectedValue(new ApiError(401, 'UNAUTHORIZED', 'Войдите'));
});

afterEach(() => {
  cleanup();
  clients.forEach((client) => client.clear());
  clients.length = 0;
  vi.restoreAllMocks();
});

describe('list cache isolation', () => {
  it('loads the full catalog independently from the sidebar and region filters', async () => {
    const waters = vi.spyOn(api, 'getWaters').mockImplementation(async (params = {}) => ({
      items: Array.from({ length: params.limit ?? 20 }, (_, index) => makeWater(`${params.region_id ?? 'all'}-${index}`)),
      next_cursor: null,
    }));
    const { wrapper } = setup();
    const { result } = renderHook(() => ({
      sidebar: useWaters({ limit: 6 }),
      catalog: useWaters({ limit: 100 }),
      regionA: useWaters({ limit: 4, region_id: 'region-a' }),
      regionB: useWaters({ limit: 4, region_id: 'region-b' }),
    }), { wrapper });

    await waitFor(() => expect(result.current.catalog.data?.items).toHaveLength(100));
    expect(result.current.sidebar.data?.items).toHaveLength(6);
    expect(result.current.regionA.data?.items[0].id).toBe('region-a-0');
    expect(result.current.regionB.data?.items[0].id).toBe('region-b-0');
    expect(waters).toHaveBeenCalledTimes(4);
  });

  it('requests distinct page sizes for feed, forum, reports and search', async () => {
    const feed = vi.spyOn(api, 'getFeed').mockResolvedValue({ items: [], next_cursor: null });
    const topics = vi.spyOn(api, 'getTopics').mockResolvedValue({ items: [], next_cursor: null });
    const reports = vi.spyOn(api, 'getReports').mockResolvedValue({ items: [], next_cursor: null });
    const search = vi.spyOn(api, 'search').mockResolvedValue({ topics: [], reports: [], waters: [], users: [] });
    const { wrapper } = setup();
    renderHook(() => {
      useFeed({ limit: 2 }); useFeed({ limit: 50 });
      useTopics({ limit: 2 }); useTopics({ limit: 50 });
      useReports({ limit: 2 }); useReports({ limit: 50 });
      useSearch({ q: 'щука', limit: 2 }); useSearch({ q: 'щука', limit: 50 });
    }, { wrapper });

    await waitFor(() => {
      expect(feed).toHaveBeenCalledTimes(2);
      expect(topics).toHaveBeenCalledTimes(2);
      expect(reports).toHaveBeenCalledTimes(2);
      expect(search).toHaveBeenCalledTimes(2);
    });
  });
});

describe('session and personal data', () => {
  it('does not request notifications for a guest', async () => {
    const notifications = vi.spyOn(api, 'getNotifications').mockResolvedValue(notificationPage('private'));
    const { wrapper } = setup();
    const { result } = renderHook(() => useNotifications(), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(result.current.data).toBeUndefined();
    expect(notifications).not.toHaveBeenCalled();
  });

  it('replaces the user and never reuses fresh notification data from the previous account', async () => {
    vi.spyOn(api, 'logout').mockResolvedValue(undefined);
    vi.spyOn(api, 'login').mockResolvedValue(undefined);
    vi.mocked(api.getMe).mockResolvedValue(makeUser('bob'));
    const notifications = vi.spyOn(api, 'getNotifications')
      .mockResolvedValueOnce(notificationPage('alice-private'))
      .mockResolvedValueOnce(notificationPage('bob-private'));
    const { client, wrapper } = setup(makeUser('alice'));
    const { result } = renderHook(() => ({ me: useMe(), notifications: useNotifications(), login: useLogin(), logout: useLogout() }), { wrapper });

    await waitFor(() => expect(result.current.notifications.data?.items[0].title).toBe('alice-private'));
    await act(async () => { await result.current.logout.mutateAsync(); });
    expect(client.getQueryData(['me'])).toBeNull();
    await waitFor(() => {
      expect(result.current.me.data).toBeNull();
      expect(result.current.notifications.data).toBeUndefined();
    });
    expect(client.getQueriesData({ queryKey: ['private'] }).every(([, data]) => data === undefined)).toBe(true);

    await act(async () => { await result.current.login.mutateAsync({ email: 'bob@example.com', password: 'a-valid-password' }); });
    await waitFor(() => expect(result.current.notifications.data?.items[0].title).toBe('bob-private'));
    expect(result.current.me.data?.id).toBe('bob');
    expect(notifications).toHaveBeenCalledTimes(2);
  });

  it('discards a previous account notification request that finishes after logout', async () => {
    const pending = deferred<NotificationPage>();
    vi.spyOn(api, 'getNotifications').mockReturnValue(pending.promise);
    vi.spyOn(api, 'logout').mockResolvedValue(undefined);
    const { client, wrapper } = setup(makeUser('alice'));
    const { result } = renderHook(() => ({ notifications: useNotifications(), logout: useLogout() }), { wrapper });
    await waitFor(() => expect(result.current.notifications.fetchStatus).toBe('fetching'));

    await act(async () => { await result.current.logout.mutateAsync(); });
    await act(async () => { pending.resolve(notificationPage('alice-private')); await pending.promise; });
    expect(result.current.notifications.data).toBeUndefined();
    expect(client.getQueriesData({ queryKey: ['private'] }).every(([, data]) => data === undefined)).toBe(true);
  });

  it('waits for the authenticated profile before resolving login', async () => {
    vi.spyOn(api, 'login').mockResolvedValue(undefined);
    const pending = deferred<User>();
    vi.mocked(api.getMe).mockReturnValue(pending.promise);
    const { client, wrapper } = setup();
    const { result } = renderHook(() => useLogin(), { wrapper });
    let finished = false;
    const completion = result.current.mutateAsync({ email: 'bob@example.com', password: 'a-valid-password' }).then(() => { finished = true; });
    await waitFor(() => expect(api.getMe).toHaveBeenCalledTimes(1));
    expect(finished).toBe(false);
    expect(client.getQueryData(['me'])).toBeNull();

    await act(async () => { pending.resolve(makeUser('bob')); await completion; });
    expect(finished).toBe(true);
    expect(client.getQueryData<User>(['me'])?.id).toBe('bob');
  });

  it('keeps registration separate from login because registration does not create a session', async () => {
    vi.spyOn(api, 'register').mockResolvedValue(undefined);
    const { client, wrapper } = setup();
    const { result } = renderHook(() => useRegister(), { wrapper });
    await act(async () => { await result.current.mutateAsync({ email: 'bob@example.com', username: 'bob', password: 'a-valid-password' }); });
    expect(client.getQueryData(['me'])).toBeNull();
    expect(api.getMe).not.toHaveBeenCalled();
  });

  it('keeps the current session when the logout request fails', async () => {
    vi.spyOn(api, 'logout').mockRejectedValue(new Error('Сеть недоступна'));
    const { client, wrapper } = setup(makeUser('alice'));
    const { result } = renderHook(() => useLogout(), { wrapper });
    await act(async () => { await expect(result.current.mutateAsync()).rejects.toThrow('Сеть недоступна'); });
    expect(client.getQueryData<User>(['me'])?.id).toBe('alice');
  });

  it('clears private data when a background session check returns 401', async () => {
    const { client, wrapper } = setup(makeUser('alice'));
    client.setQueryData(['private', 'alice', 'notifications'], notificationPage('alice-private'));
    const { result } = renderHook(() => useMe(), { wrapper });
    expect(result.current.data?.id).toBe('alice');
    await act(async () => {
      const refreshed = await result.current.refetch();
      expect(refreshed.error).toBeNull();
      expect(refreshed.data).toBeNull();
    });
    await waitFor(() => expect(result.current.data).toBeNull());
    expect(client.getQueryData(['private', 'alice', 'notifications'])).toBeUndefined();
  });

  it('does not let a cancelled profile request erase the next account personal cache', async () => {
    const staleProfile = deferred<User>();
    let staleSignal: AbortSignal | undefined;
    vi.mocked(api.getMe).mockImplementationOnce((signal) => {
      staleSignal = signal;
      return staleProfile.promise;
    }).mockResolvedValue(makeUser('bob'));
    vi.spyOn(api, 'getNotifications')
      .mockResolvedValueOnce(notificationPage('alice-private'))
      .mockResolvedValueOnce(notificationPage('bob-private'));
    vi.spyOn(api, 'logout').mockResolvedValue(undefined);
    vi.spyOn(api, 'login').mockResolvedValue(undefined);
    const { client, wrapper } = setup(makeUser('alice'));
    const { result } = renderHook(() => ({ me: useMe(), notifications: useNotifications(), login: useLogin(), logout: useLogout() }), { wrapper });
    await waitFor(() => expect(result.current.notifications.data?.items[0].title).toBe('alice-private'));
    let oldRequest!: Promise<unknown>;
    act(() => { oldRequest = result.current.me.refetch(); });
    await waitFor(() => expect(staleSignal).toBeDefined());

    await act(async () => { await result.current.logout.mutateAsync(); await oldRequest; });
    await act(async () => { await result.current.login.mutateAsync({ email: 'bob@example.com', password: 'a-valid-password' }); });
    await waitFor(() => expect(result.current.notifications.data?.items[0].title).toBe('bob-private'));
    const bobCache = client.getQueriesData({ queryKey: ['private', 'bob'] });
    expect(bobCache.length).toBeGreaterThan(0);
    expect(staleSignal?.aborted).toBe(true);

    await act(async () => { staleProfile.resolve(makeUser('alice')); await staleProfile.promise; });
    expect(client.getQueryData<User>(['me'])?.id).toBe('bob');
    expect(client.getQueriesData({ queryKey: ['private', 'bob'] })).toEqual(bobCache);
  });
});
