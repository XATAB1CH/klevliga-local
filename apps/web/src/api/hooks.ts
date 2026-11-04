import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, ApiError, type FeedParams, type ReportsParams, type SearchParams, type TopicsParams, type WatersParams } from './client';

export function useFeed(params: FeedParams = {}) {
  return useQuery({
    queryKey: ['feed', params.kind ?? 'all', params.cursor ?? null],
    queryFn: () => api.getFeed(params),
    placeholderData: (prev) => prev,
  });
}

export function useCategories() {
  return useQuery({ queryKey: ['categories'], queryFn: api.getCategories, staleTime: 5 * 60_000 });
}

export function useTopics(params: TopicsParams = {}) {
  return useQuery({
    queryKey: ['topics', params.category_id ?? null, params.cursor ?? null],
    queryFn: () => api.getTopics(params),
    placeholderData: (prev) => prev,
  });
}

export function useTopic(id: string) {
  return useQuery({ queryKey: ['topic', id], queryFn: () => api.getTopic(id), enabled: !!id });
}

export function usePosts(topicId: string) {
  return useQuery({ queryKey: ['posts', topicId], queryFn: () => api.getPosts(topicId), enabled: !!topicId });
}

export function useReports(params: ReportsParams = {}) {
  return useQuery({
    queryKey: ['reports', params.water_id ?? null, params.cursor ?? null],
    queryFn: () => api.getReports(params),
    placeholderData: (prev) => prev,
  });
}

export function useWaters(params: WatersParams = {}) {
  return useQuery({
    queryKey: ['waters', params.q ?? null, params.cursor ?? null],
    queryFn: () => api.getWaters(params),
    placeholderData: (prev) => prev,
  });
}

export function useWater(id: string) {
  return useQuery({ queryKey: ['water', id], queryFn: () => api.getWater(id), enabled: !!id });
}

export function useSearch(params: SearchParams) {
  return useQuery({
    queryKey: ['search', params.q, params.type ?? 'all'],
    queryFn: () => api.search(params),
    enabled: params.q.trim().length >= 2,
  });
}

// Возвращает текущего пользователя или null, если сессии нет.
export function useMe() {
  return useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      try {
        return await api.getMe();
      } catch (e) {
        if (e instanceof ApiError && e.status === 401) return null;
        throw e;
      }
    },
    retry: false,
    staleTime: 60_000,
  });
}

export function useNotifications() {
  return useQuery({ queryKey: ['notifications'], queryFn: () => api.getNotifications() });
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
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useRegister() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (p: { email: string; username: string; password: string }) => api.register(p.email, p.username, p.password),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useLogout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.logout(),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['me'] }),
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
