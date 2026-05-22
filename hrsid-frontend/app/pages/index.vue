<script setup lang="ts">
definePageMeta({ middleware: 'auth', layout: 'default' })

interface Application {
  id: number
  name: string
  slug: string
  base_url: string
  icon_url: string | null
  description: string
}

const { apiFetch } = useApi()

const { data, status } = await useAsyncData('my-apps', () =>
  apiFetch<{ applications: Application[] }>('/api/v1/my-apps')
)

const apps = computed(() => data.value?.applications ?? [])
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Portal Aplikasi</h1>
      <p class="text-muted mt-1">Akses semua aplikasi yang tersedia untuk departemen Anda</p>
    </div>

    <div v-if="status === 'pending'" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <USkeleton v-for="i in 4" :key="i" class="h-40 rounded-xl" />
    </div>

    <UAlert
      v-else-if="apps.length === 0"
      color="neutral"
      variant="soft"
      icon="i-lucide-inbox"
      title="Belum ada aplikasi"
      description="Belum ada aplikasi yang tersedia untuk departemen Anda. Hubungi administrator."
    />

    <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <a
        v-for="application in apps"
        :key="application.id"
        :href="application.base_url"
        target="_blank"
        rel="noopener noreferrer"
        class="group block"
      >
        <UCard class="h-full transition hover:ring-2 hover:ring-primary cursor-pointer">
          <div class="flex flex-col items-center text-center gap-3 py-2">
            <div
              class="size-14 rounded-xl bg-primary/10 flex items-center justify-center overflow-hidden"
            >
              <img
                v-if="application.icon_url"
                :src="application.icon_url"
                :alt="application.name"
                class="size-10 object-contain"
              />
              <UIcon v-else name="i-lucide-app-window" class="size-8 text-primary" />
            </div>
            <div>
              <p class="font-semibold group-hover:text-primary transition-colors">
                {{ application.name }}
              </p>
              <p class="text-xs text-muted mt-0.5 line-clamp-2">{{ application.description }}</p>
            </div>
          </div>
        </UCard>
      </a>
    </div>
  </div>
</template>
