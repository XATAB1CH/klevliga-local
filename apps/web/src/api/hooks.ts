import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { api, ApiError, type FeedParams, type ReportsParams, type SearchParams, type TopicsParams, type WatersParams } from './client';
import type { User } from './types';

type CursorParams = { cursor?: string; limit?: number };

// Every parameter affecting a response must also identify its cache entry.
function listParams<T extends CursorParams>(params: T) {
  return { ...params, cursor: params.cursor || null, limit: params.limit ?? 20 };
}

async function clearPrivateQueries(client: QueryClient) {
  await client.cancelQueries({ queryKey: ['private'] });
  client.removeQueries({ queryKey: ['private'] });
}

async function clearSession(client: QueryClient) {
  await client.cancelQueries({ queryKey: ['me'] });
  client.setQueryData(['me'], null);
  await clearPrivateQueries(client);
}

async function loadCurrentUser(client: QueryClient) {
  let user: User | null;
  try {
    user = await api.getMe();
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) throw error;
    user = null;
  }
  const previous = client.getQueryData<User | null>(['me']);
  if (previous?.id !== user?.id) await clearPrivateQueries(client);
  return user;
}

export function useFeed(params: FeedParams = {}) {
  return useQuery({
    queryKey: ['feed', { ...listParams(params), kind: params.kind ?? 'all' }],
    queryFn: () => api.getFeed(params),
    placeholderData: (prev) => prev,
  });
}

export function useCategories() {
  return useQuery({ queryKey: ['categories'], queryFn: api.getCategories, staleTime: 5 * 60_000 });
}

export function useTopics(params: TopicsParams = {}) {
  return useQuery({
    queryKey: ['topics', listParams(params)],
    queryFn: () => api.getTopics(params),
    placeholderData: (prev) => prev,
  });
}

export function useTopic(id: string) {
  return useQuery({ queryKey: ['topic', id], queryFn: () => api.getTopic(id), enabled: !!id });
}

export function usePosts(topicId: string, params: CursorParams = {}) {
  return useQuery({ queryKey: ['posts', topicId, listParams(params)], queryFn: () => api.getPosts(topicId, params), enabled: !!topicId });
}

export function useReports(params: ReportsParams = {}) {
  return useQuery({
    queryKey: ['reports', listParams(params)],
    queryFn: () => api.getReports(params),
    placeholderData: (prev) => prev,
  });
}

export function useWaters(params: WatersParams = {}) {
  return useQuery({
    queryKey: ['waters', listParams(params)],
    queryFn: () => api.getWaters(params),
    placeholderData: (prev) => prev,
  });
}

export function useWater(id: string) {
  return useQuery({ queryKey: ['water', id], queryFn: () => api.getWater(id), enabled: !!id });
}

export function useSearch(params: SearchParams) {
  return useQuery({
    queryKey: ['search', { ...params, type: params.type ?? 'all', limit: params.limit ?? 20 }],
    queryFn: () => api.search(params),
    enabled: params.q.trim().length >= 2,
  });
}

// Возвращает текущего пользователя или null, если сессии нет.
export function useMe() {
  const client = useQueryClient();
  return useQuery({
    queryKey: ['me'],
    queryFn: () => loadCurrentUser(client),
    retry: false,
    staleTime: 60_000,
  });
}

export function useNotifications(params: CursorParams = {}) {
  const { data: user } = useMe();
  return useQuery({
    queryKey: ['private', user?.id ?? null, 'notifications', listParams(params)],
    queryFn: () => api.getNotifications(params),
    enabled: !!user,
  });
}

function useInvalidate() {
  const qc = useQueryClient();
  return (keys: string[][]) => {
    for (const key of keys) void qc.invalidateQueries({ queryKey: key });
  };
}

export function useLogin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (p: { email: string; password: string }) => api.login(p.email, p.password),
    onSuccess: async () => {
      await clearSession(qc);
      // Resolve login only after the authenticated profile is available.
      await qc.fetchQuery({ queryKey: ['me'], queryFn: api.getMe, staleTime: 0 });
    },
  });
}

export function useRegister() {
  return useMutation({
    mutationFn: (p: { email: string; username: string; password: string }) => api.register(p.email, p.username, p.password),
    // Registration creates an account, but the API does not create a session.
  });
}

export function useLogout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.logout(),
    onSuccess: () => clearSession(qc),
  });
}

export function useCreateTopic() {
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (p: { categoryId: string; title: string; body: string }) => api.createTopic(p.categoryId, p.title, p.body),
    onSuccess: () => invalidate([['feed'], ['topics']]),
  });
}

export function useCreateReport() {
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (p: { title: string; body: string; water_id?: string | null; fished_at: string; location_visibility: string; species: string[] }) => api.createReport(p),
    onSuccess: () => invalidate([['feed'], ['reports']]),
  });
}

export function useReply(topicId: string) {
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (p: { body: string; parentPostId?: string | null }) => api.reply(topicId, p.body, p.parentPostId),
    onSuccess: () => invalidate([['posts', topicId], ['topic', topicId], ['feed'], ['topics']]),
  });
}

export function useFollow() {
  return useMutation({
    mutationFn: (p: { targetType: string; targetId: string }) => api.follow(p.targetType, p.targetId),
  });
}

export function useUnfollow() {
  return useMutation({
    mutationFn: (p: { targetType: string; targetId: string }) => api.unfollow(p.targetType, p.targetId),
  });
}

export function useFlag() {
  return useMutation({
    mutationFn: (p: { targetType: string; targetId: string; reason: string }) => api.flag(p.targetType, p.targetId, p.reason),
  });
}
