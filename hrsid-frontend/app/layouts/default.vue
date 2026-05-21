<script setup lang="ts">
const auth = useAuthStore();

const links = computed(() => {
  const base = [{ label: 'Dashboard', to: '/', icon: 'i-lucide-layout-dashboard' }];
  if (auth.isAdmin) {
    base.push({ label: 'Manajemen User', to: '/admin/users', icon: 'i-lucide-users' });
  }
  return base;
});
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
    <UHeader>
      <template #left>
        <NuxtLink
          to="/"
          class="flex items-center gap-2 font-bold text-lg text-primary"
        >
          <UIcon
            name="i-lucide-shield-check"
            class="size-6"
          />
          HRSID
        </NuxtLink>
      </template>

      <template #body>
        <UNavigationMenu
          :links="links"
          class="hidden md:flex"
        />
      </template>

      <template #right>
        <div class="flex items-center gap-3">
          <span class="text-sm text-muted hidden sm:block">{{ auth.user?.fullName }}</span>
          <UButton
            label="Logout"
            icon="i-lucide-log-out"
            color="neutral"
            variant="ghost"
            size="sm"
            @click="auth.logout()"
          />
        </div>
      </template>
    </UHeader>

    <UMain class="flex-1">
      <UContainer class="py-8">
        <slot />
      </UContainer>
    </UMain>
  </div>
</template>
