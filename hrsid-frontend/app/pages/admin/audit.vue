<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'admin'], layout: 'default' })

interface AuditLog {
  id: string
  public_id: string
  action: string
  actor_user_id?: string
  actor_identifier?: string
  actor_full_name?: string
  actor_nik?: string
  target_user_id?: string
  target_full_name?: string
  target_nik?: string
  ip_address: string
  user_agent: string
  metadata?: Record<string, unknown>
  created_at: string
}

const { apiFetch } = useApi()

const { data, refresh, status } = await useAsyncData('admin-audit-logs', () =>
  apiFetch<{ logs: AuditLog[] }>('/api/v1/admin/audit-logs')
)
const logs = computed(() => data.value?.logs ?? [])

const search = ref('')
const actionFilter = ref<string>('all')

const actionOptions = [
  { label: 'Semua Aksi', value: 'all' },
  { label: 'Login Sukses', value: 'login.success' },
  { label: 'Login Gagal', value: 'login.failure' },
  { label: 'Logout', value: 'logout' },
  { label: 'Refresh Token', value: 'token.refresh' },
  { label: 'Pembuatan User', value: 'user.create' },
  { label: 'Nonaktifkan User', value: 'user.deactivate' },
  { label: 'Aktifkan User', value: 'user.reactivate' },
  { label: 'Cabut Sesi', value: 'session.revoke' }
]

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  return logs.value.filter((l) => {
    if (actionFilter.value !== 'all' && l.action !== actionFilter.value) return false
    if (!q) return true
    return (
      l.action.toLowerCase().includes(q) ||
      (l.actor_full_name ?? '').toLowerCase().includes(q) ||
      (l.actor_identifier ?? '').toLowerCase().includes(q) ||
      (l.actor_nik ?? '').toLowerCase().includes(q) ||
      (l.target_full_name ?? '').toLowerCase().includes(q) ||
      (l.target_nik ?? '').toLowerCase().includes(q) ||
      l.ip_address.toLowerCase().includes(q)
    )
  })
})

const columns = [
  { accessorKey: 'created_at', header: 'Waktu' },
  { accessorKey: 'action', header: 'Aksi' },
  { accessorKey: 'actor', header: 'Pelaku' },
  { accessorKey: 'target', header: 'Target' },
  { accessorKey: 'ip_address', header: 'IP' },
  { accessorKey: 'user_agent', header: 'User Agent' },
  { accessorKey: 'metadata', header: 'Detail' }
]

function fmtDate(s: string) {
  if (!s) return '-'
  try {
    return new Date(s).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'medium' })
  } catch {
    return s
  }
}

function actionColor(a: string): 'success' | 'error' | 'warning' | 'info' | 'neutral' {
  if (a === 'login.success') return 'success'
  if (a === 'login.failure') return 'error'
  if (a === 'user.deactivate' || a === 'session.revoke') return 'warning'
  if (a === 'user.reactivate' || a === 'user.create') return 'info'
  return 'neutral'
}

function summarizeUA(ua: string) {
  if (!ua) return '-'
  if (ua.length > 60) return ua.slice(0, 57) + '…'
  return ua
}

function actorLabel(l: AuditLog) {
  if (l.actor_full_name) return `${l.actor_full_name}${l.actor_nik ? ` (${l.actor_nik})` : ''}`
  if (l.actor_identifier) return l.actor_identifier
  return 'Anonim'
}

function targetLabel(l: AuditLog) {
  if (l.target_full_name) return `${l.target_full_name}${l.target_nik ? ` (${l.target_nik})` : ''}`
  return '—'
}

function metaPreview(m?: Record<string, unknown>) {
  if (!m) return '—'
  const entries = Object.entries(m)
  if (entries.length === 0) return '—'
  return entries.map(([k, v]) => `${k}=${typeof v === 'string' ? v : JSON.stringify(v)}`).join(', ')
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold">Audit Log</h1>
        <p class="text-muted mt-1">Jejak akses dan aksi pengguna pada sistem HRSID.</p>
      </div>
      <UButton
        label="Refresh"
        icon="i-lucide-refresh-cw"
        color="neutral"
        variant="soft"
        :loading="status === 'pending'"
        @click="refresh()"
      />
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col sm:flex-row gap-3">
          <UInput
            v-model="search"
            placeholder="Cari NIK, nama, IP, atau aksi..."
            icon="i-lucide-search"
            class="flex-1"
          />
          <USelect v-model="actionFilter" :items="actionOptions" class="w-full sm:w-56" />
        </div>
      </template>

      <UTable :data="filtered" :columns="columns" :loading="status === 'pending'">
        <template #created_at-cell="{ row }">
          <span class="text-sm whitespace-nowrap">{{ fmtDate(row.original.created_at) }}</span>
        </template>

        <template #action-cell="{ row }">
          <UBadge :color="actionColor(row.original.action)" variant="soft" :label="row.original.action" />
        </template>

        <template #actor-cell="{ row }">
          <span class="text-sm">{{ actorLabel(row.original) }}</span>
        </template>

        <template #target-cell="{ row }">
          <span class="text-sm text-muted">{{ targetLabel(row.original) }}</span>
        </template>

        <template #ip_address-cell="{ row }">
          <span class="font-mono text-xs">{{ row.original.ip_address || '-' }}</span>
        </template>

        <template #user_agent-cell="{ row }">
          <span class="text-xs text-muted">{{ summarizeUA(row.original.user_agent) }}</span>
        </template>

        <template #metadata-cell="{ row }">
          <span class="text-xs text-muted">{{ metaPreview(row.original.metadata) }}</span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>
