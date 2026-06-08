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
const toast = useToast()

const { data, status } = await useAsyncData('my-apps', () =>
  apiFetch<{ applications: Application[] }>('/api/v1/my-apps')
)

const apps = computed(() => data.value?.applications ?? [])

const launchingSlug = ref<string | null>(null)

async function launchApp(application: Application) {
  if (launchingSlug.value) return
  launchingSlug.value = application.slug

  // Open the tab synchronously so the browser doesn't block it as a popup
  // after the async launch request resolves. Omit the features arg, otherwise
  // browsers open a popup window instead of a new tab.
  const win = window.open('', '_blank')

  try {
    const { redirect_url } = await apiFetch<{ redirect_url: string }>(
      `/api/v1/launch/${application.slug}`,
      { method: 'POST' }
    )

    if (win) {
      win.opener = null
      win.location.href = redirect_url
    } else {
      // Popup was blocked; fall back to navigating the current tab.
      window.location.href = redirect_url
    }
  } catch {
    win?.close()
    toast.add({
      title: 'Gagal membuka aplikasi',
      description: `Tidak dapat memulai sesi SSO untuk ${application.name}. Silakan coba lagi.`,
      color: 'error',
      icon: 'i-lucide-alert-triangle',
    })
  } finally {
    launchingSlug.value = null
  }
}
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
      <button
        v-for="application in apps"
        :key="application.id"
        type="button"
        :disabled="launchingSlug !== null"
        class="group block text-left disabled:cursor-not-allowed"
        @click="launchApp(application)"
      >
        <UCard class="h-full transition hover:ring-2 hover:ring-primary cursor-pointer">
          <div class="flex flex-col items-center text-center gap-3 py-2">
            <div
              class="size-14 rounded-xl bg-primary/10 flex items-center justify-center overflow-hidden"
            >
              <UIcon
                v-if="launchingSlug === application.slug"
                name="i-lucide-loader-circle"
                class="size-8 text-primary animate-spin"
              />
              <img
                v-else-if="application.icon_url"
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
      </button>
    </div>
  </div>
</template>
