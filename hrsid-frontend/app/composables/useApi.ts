import type { FetchOptions } from 'ofetch';

export function useApi() {
  const auth = useAuthStore();
  const config = useRuntimeConfig();

  async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
    const authHeaders: Record<string, string> = auth.accessToken
      ? { Authorization: `Bearer ${auth.accessToken}` }
      : {};

    try {
      return await $fetch<T>(path, {
        baseURL: config.public.apiBase,
        credentials: 'include',
        headers: { ...authHeaders, ...(options.headers as Record<string, string> | undefined) },
        ...options
      });
    } catch (err: unknown) {
      const fetchErr = err as { response?: { status?: number } };
      if (fetchErr.response?.status === 401) {
        try {
          await auth.refreshToken();
          return await $fetch<T>(path, {
            baseURL: config.public.apiBase,
            credentials: 'include',
            headers: {
              Authorization: `Bearer ${auth.accessToken}`,
              ...(options.headers as Record<string, string> | undefined)
            },
            ...options
          });
        } catch {
          auth.clearAuth();
          await navigateTo('/login');
          throw err;
        }
      }
      throw err;
    }
  }

  return { apiFetch };
}
