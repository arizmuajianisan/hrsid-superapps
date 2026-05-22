<script setup lang="ts">
definePageMeta({ middleware: ['auth'], layout: 'default' })

interface Session {
  id: string
  user_id: string
  ip_address: string
  user_agent: string
  is_blocked: boolean
  expiry: string
  created_at: string
}

interface SessionsResponse {
  sessions: Session[]
  current_session_id: string
}

const { apiFetch } = useApi()
const toast = useToast()

const { data, refresh, status } = await useAsyncData('my-sessions', () =>
  apiFetch<SessionsResponse>('/api/v1/sessions')
)
const sessions = computed(() => data.value?.sessions ?? [])
const currentSessionId = computed(() => data.value?.current_session_id ?? '')

const columns = [
  { accessorKey: 'created_at', header: 'Dibuat' },
  { accessorKey: 'ip_address', header: 'IP Address' },
  { accessorKey: 'user_agent', header: 'Perangkat / Browser' },
  { accessorKey: 'expiry', header: 'Berlaku Hingga' },
  { accessorKey: 'is_blocked', header: 'Status' },
  { id: 'actions', header: 'Aksi' }
]

function fmtDate(s: string) {
  if (!s) return '-'
  try {
    return new Date(s).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return s
  }
}

function summarizeUA(ua: string) {
  if (!ua) return 'Tidak diketahui'
  if (ua.length > 80) return ua.slice(0, 77) + '…'
  return ua
}

const actionLoading = ref<string | null>(null)
const confirmModal = reactive({ open: false, session: null as Session | null })

function openConfirm(s: Session) {
  confirmModal.session = s
  confirmModal.open = true
}
function closeConfirm() {
  confirmModal.open = false
  confirmModal.session = null
}

async function revoke(id: string) {
  actionLoading.value = id
  try {
    await apiFetch(`/api/v1/sessions/${id}/revoke`, { method: 'POST' })
    toast.add({ title: 'Berhasil', description: 'Sesi telah dicabut.', color: 'success' })
    await refresh()
  } catch {
    toast.add({ title: 'Gagal', description: 'Tidak dapat mencabut sesi.', color: 'error' })
  } finally {
    actionLoading.value = null
  }
}

async function handleConfirmRevoke() {
  const s = confirmModal.session
  closeConfirm()
  if (s) await revoke(s.id)
}
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Sesi Saya</h1>
      <p class="text-muted mt-1">Kelola perangkat yang masuk ke akun Anda. Cabut sesi yang tidak dikenali.</p>
    </div>

    <UCard>
      <UTable :data="sessions" :columns="columns" :loading="status === 'pending'">
        <template #created_at-cell="{ row }">
          {{ fmtDate(row.original.created_at) }}
        </template>

        <template #ip_address-cell="{ row }">
          <span class="font-mono text-sm">{{ row.original.ip_address || '-' }}</span>
        </template>

        <template #user_agent-cell="{ row }">
          <span class="text-sm text-muted">{{ summarizeUA(row.original.user_agent) }}</span>
        </template>

        <template #expiry-cell="{ row }">
          {{ fmtDate(row.original.expiry) }}
        </template>

        <template #is_blocked-cell="{ row }">
          <UBadge
            v-if="row.original.id === currentSessionId"
            color="primary"
            variant="soft"
            label="Sesi Saat Ini"
          />
          <UBadge
            v-else-if="row.original.is_blocked"
            color="neutral"
            variant="soft"
            label="Dicabut"
          />
          <UBadge v-else color="success" variant="soft" label="Aktif" />
        </template>

        <template #actions-cell="{ row }">
          <UButton
            v-if="!row.original.is_blocked && row.original.id !== currentSessionId"
            label="Cabut"
            color="error"
            variant="soft"
            size="xs"
            :loading="actionLoading === row.original.id"
            @click="openConfirm(row.original)"
          />
          <span v-else class="text-xs text-muted">—</span>
        </template>
      </UTable>
    </UCard>

    <UModal
      v-model:open="confirmModal.open"
      title="Cabut Sesi"
      :ui="{ footer: 'justify-end' }"
    >
      <template #body>
        <div class="flex items-start gap-4">
          <div class="flex items-center justify-center rounded-full p-2 shrink-0 bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
            <UIcon name="i-lucide-log-out" class="w-5 h-5" />
          </div>
          <div>
            <p class="font-medium text-sm text-highlighted">Cabut sesi ini?</p>
            <p class="text-sm text-muted mt-1">
              Perangkat dengan IP
              <span class="font-mono">{{ confirmModal.session?.ip_address || '-' }}</span>
              akan langsung kehilangan akses.
            </p>
          </div>
        </div>
      </template>

      <template #footer>
        <UButton label="Batal" color="neutral" variant="outline" @click="closeConfirm" />
        <UButton
          label="Ya, Cabut"
          color="error"
          :loading="!!actionLoading"
          @click="handleConfirmRevoke"
        />
      </template>
    </UModal>
  </div>
</template>
