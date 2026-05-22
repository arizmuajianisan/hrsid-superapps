import { defineStore } from "pinia";

interface AuthUser {
  fullName: string;
  role: string;
}

interface LoginResponse {
  access_token: string;
  user: {
    full_name: string;
    role: string;
  };
}

interface RefreshResponse {
  access_token: string;
}

interface MeResponse {
  user: {
    full_name: string;
    role: string;
  };
}

export const useAuthStore = defineStore("auth", () => {
  const config = useRuntimeConfig();
  const accessToken = ref<string | null>(null);
  const user = ref<AuthUser | null>(null);

  const isAuthenticated = computed(() => !!accessToken.value);
  const isAdmin = computed(() => user.value?.role === "admin");

  let refreshPromise: Promise<void> | null = null;

  async function login(identifier: string, password: string) {
    const data = await $fetch<LoginResponse>("/api/v1/login", {
      baseURL: config.public.apiBase,
      method: "POST",
      credentials: "include",
      body: { identifier, password }
    });
    accessToken.value = data.access_token;
    user.value = { fullName: data.user.full_name, role: data.user.role };
  }

  async function _doRefresh() {
    const data = await $fetch<RefreshResponse>("/api/v1/refresh", {
      baseURL: config.public.apiBase,
      method: "POST",
      credentials: "include"
    });
    accessToken.value = data.access_token;
  }

  async function refreshToken() {
    if (refreshPromise) return refreshPromise;
    refreshPromise = _doRefresh().finally(() => {
      refreshPromise = null;
    });
    return refreshPromise;
  }

  async function initAuth() {
    try {
      await _doRefresh();
      const data = await $fetch<MeResponse>("/api/v1/me", {
        baseURL: config.public.apiBase,
        credentials: "include",
        headers: { Authorization: `Bearer ${accessToken.value}` }
      });
      user.value = { fullName: data.user.full_name, role: data.user.role };
    } catch (err) {
      console.warn("[initAuth] Session restore failed:", err);
      clearAuth();
    }
  }

  async function logout() {
    await $fetch("/api/v1/logout", {
      baseURL: config.public.apiBase,
      method: "POST",
      credentials: "include"
    }).catch(() => {});
    clearAuth();
    await navigateTo("/login");
  }

  function clearAuth() {
    accessToken.value = null;
    user.value = null;
  }

  return {
    accessToken,
    user,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    refreshToken,
    clearAuth,
    initAuth
  };
});
