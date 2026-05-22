<script setup lang="ts">
const auth = useAuthStore()
const { isAdmin, user } = storeToRefs(auth)

const userMenuItems = computed(() => {
  const groups = []

  groups.push([
    {
      label: 'Sesi Saya',
      icon: 'i-lucide-monitor',
      to: '/sessions'
    }
  ])

  if (isAdmin.value) {
    groups.push([
      {
        label: 'Manajemen User',
        icon: 'i-lucide-users',
        to: '/admin/users'
      },
      {
        label: 'Audit Log',
        icon: 'i-lucide-scroll-text',
        to: '/admin/audit'
      }
    ])
  }

  groups.push([
    {
      label: 'Logout',
      icon: 'i-lucide-log-out',
      color: 'error' as const,
      onSelect: () => auth.logout()
    }
  ])

  return groups
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
    <UHeader>
      <template #left>
        <NuxtLink to="/" class="flex items-center gap-2 font-bold text-lg text-primary">
          <UIcon name="i-lucide-shield-check" class="size-6" />
          HRSID
        </NuxtLink>
      </template>

      <template #right>
        <UDropdownMenu :items="userMenuItems">
          <UButton
            color="neutral"
            variant="ghost"
            trailing-icon="i-lucide-chevron-down"
            class="gap-2"
          >
            <div class="text-left hidden sm:block">
              <p class="text-sm font-medium leading-none">
                {{ user?.fullName }}
              </p>
              <p class="text-xs text-muted mt-0.5 capitalize">{{ user?.role }} {{ user?.dept }}</p>
            </div>
            <UIcon name="i-lucide-user-circle" class="size-5 sm:hidden" />
          </UButton>
        </UDropdownMenu>
      </template>
    </UHeader>

    <UMain class="flex-1">
      <UContainer class="py-8">
        <slot />
      </UContainer>
    </UMain>
  </div>
</template>
