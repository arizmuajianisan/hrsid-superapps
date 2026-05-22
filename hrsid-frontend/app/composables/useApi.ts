import type { FetchOptions } from "ofetch";
import type { NitroFetchRequest } from "nitropack";

type HttpMethod
  = | "GET"
    | "HEAD"
    | "PATCH"
    | "POST"
    | "PUT"
    | "DELETE"
    | "TRACE"
    | "CONNECT"
    | "OPTIONS"
    | "TRACE"
    | "get"
    | "head"
    | "patch"
    | "post"
    | "put"
    | "delete"
    | "connect"
    | "options"
    | "trace";

export function useApi() {
  const auth = useAuthStore();
  const config = useRuntimeConfig();

  function buildFetchOptions(
    authHeaders: Record<string, string>,
    options: FetchOptions
  ) {
    return {
      baseURL: config.public.apiBase as string,
      credentials: "include" as RequestCredentials,
      headers: {
        ...authHeaders,
        ...(options.headers as Record<string, string> | undefined)
      },
      ...options,
      method: options.method as HttpMethod | undefined
    };
  }

  async function apiFetch<T>(
    path: NitroFetchRequest,
    options: FetchOptions = {}
  ): Promise<T> {
    const authHeaders: Record<string, string> = auth.accessToken
      ? { Authorization: `Bearer ${auth.accessToken}` }
      : {};

    try {
      return await $fetch<T>(path, buildFetchOptions(authHeaders, options));
    } catch (err: unknown) {
      const fetchErr = err as { response?: { status?: number } };
      if (fetchErr.response?.status === 401) {
        try {
          await auth.refreshToken();
          const refreshedHeaders: Record<string, string> = {
            Authorization: `Bearer ${auth.accessToken}`,
            ...(options.headers as Record<string, string> | undefined)
          };
          return await $fetch<T>(
            path,
            buildFetchOptions(refreshedHeaders, options)
          );
        } catch {
          auth.clearAuth();
          await navigateTo("/login");
          throw err;
        }
      }
      throw err;
    }
  }

  return { apiFetch };
}
