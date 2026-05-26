<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'admin'], layout: 'default' })

interface Department {
  id: number
  name: string
}

interface Application {
  id: number
  name: string
  slug: string
  base_url: string
  icon_url: string | null
  description: string
  department_ids: number[]
}

const { apiFetch } = useApi()
const toast = useToast()

// --- Data ---
const {
  data: appsData,
  refresh,
  status
} = await useAsyncData('admin-applications', () =>
  apiFetch<{ applications: Application[] }>('/api/v1/admin/applications')
)
const { data: deptsData } = await useAsyncData('admin-departments', () =>
  apiFetch<{ departments: Department[] }>('/api/v1/admin/departments')
)

const applications = computed(() => appsData.value?.applications ?? [])
const departments = computed(() => deptsData.value?.departments ?? [])

function deptName(id: number) {
  return departments.value.find((d) => d.id === id)?.name ?? `Dept ${id}`
}

// --- Table columns ---
const columns = [
  { accessorKey: 'name', header: 'Nama Aplikasi' },
  { accessorKey: 'slug', header: 'Slug' },
  { accessorKey: 'base_url', header: 'URL' },
  { accessorKey: 'description', header: 'Deskripsi' },
  { id: 'departments', header: 'Departemen' },
  { id: 'actions', header: 'Aksi' }
]

// --- Form state ---
const isModalOpen = ref(false)
const isEditing = ref(false)
const formLoading = ref(false)
const formError = ref<string | null>(null)

const form = reactive({
  id: 0,
  name: '',
  slug: '',
  base_url: '',
  icon_url: '',
  description: '',
  department_ids: [] as number[]
})

function resetForm() {
  formError.value = null
  Object.assign(form, {
    id: 0,
    name: '',
    slug: '',
    base_url: '',
    icon_url: '',
    description: '',
    department_ids: []
  })
}

function openCreate() {
  isEditing.value = false
  resetForm()
  isModalOpen.value = true
}

function openEdit(app: Application) {
  isEditing.value = true
  formError.value = null
  Object.assign(form, {
    id: app.id,
    name: app.name,
    slug: app.slug,
    base_url: app.base_url,
    icon_url: app.icon_url ?? '',
    description: app.description,
    department_ids: [...app.department_ids]
  })
  isModalOpen.value = true
}

function toggleDept(id: number) {
  const idx = form.department_ids.indexOf(id)
  if (idx === -1) form.department_ids.push(id)
  else form.department_ids.splice(idx, 1)
}

async function handleSubmit() {
  formError.value = null
  formLoading.value = true
  const body = {
    name: form.name,
    slug: form.slug,
    base_url: form.base_url,
    icon_url: form.icon_url || null,
    description: form.description,
    department_ids: form.department_ids
  }
  try {
    if (isEditing.value) {
      await apiFetch(`/api/v1/admin/applications/${form.id}`, { method: 'PUT', body })
      toast.add({
        title: 'Berhasil',
        description: 'Aplikasi berhasil diperbarui.',
        color: 'success'
      })
    } else {
      await apiFetch('/api/v1/admin/applications', { method: 'POST', body })
      toast.add({
        title: 'Berhasil',
        description: 'Aplikasi berhasil ditambahkan.',
        color: 'success'
      })
    }
    isModalOpen.value = false
    await refresh()
  } catch (err: unknown) {
    const fe = err as { response?: { status?: number } }
    if (fe.response?.status === 409) {
      formError.value = 'Slug sudah digunakan aplikasi lain.'
    } else {
      formError.value = 'Terjadi kesalahan. Silakan coba lagi.'
    }
  } finally {
    formLoading.value = false
  }
}

// --- Delete ---
const deleteModal = reactive({ open: false, app: null as Application | null })
const deleteLoading = ref(false)

function openDelete(app: Application) {
  deleteModal.app = app
  deleteModal.open = true
}

async function handleDelete() {
  if (!deleteModal.app) return
  const target = deleteModal.app
  deleteModal.open = false
  deleteLoading.value = true
  try {
    await apiFetch(`/api/v1/admin/applications/${target.id}`, { method: 'DELETE' })
    toast.add({
      title: 'Berhasil',
      description: `Aplikasi "${target.name}" dihapus.`,
      color: 'success'
    })
    await refresh()
  } catch {
    toast.add({ title: 'Gagal', description: 'Terjadi kesalahan saat menghapus.', color: 'error' })
  } finally {
    deleteLoading.value = false
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold">Manajemen Aplikasi</h1>
        <p class="text-muted mt-1">Kelola aplikasi yang tersedia di portal HRSID</p>
      </div>
      <UButton label="Tambah Aplikasi" icon="i-lucide-plus" @click="openCreate" />
    </div>

    <UCard>
      <UTable :data="applications" :columns="columns" :loading="status === 'pending'">
        <template #base_url-cell="{ row }">
          <a
            :href="row.original.base_url"
            target="_blank"
            class="text-primary hover:underline truncate max-w-[180px] block"
          >
            {{ row.original.base_url }}
          </a>
        </template>

        <template #description-cell="{ row }">
          <span class="truncate max-w-[160px] block text-muted">{{
            row.original.description
          }}</span>
        </template>

        <template #departments-cell="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge
              v-for="id in row.original.department_ids"
              :key="id"
              color="primary"
              variant="soft"
              size="xs"
              :label="deptName(id)"
            />
            <span v-if="row.original.department_ids.length === 0" class="text-muted text-sm"
              >—</span
            >
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              label="Edit"
              color="neutral"
              variant="soft"
              size="xs"
              icon="i-lucide-pencil"
              @click="openEdit(row.original)"
            />
            <UButton
              label="Hapus"
              color="error"
              variant="soft"
              size="xs"
              icon="i-lucide-trash-2"
              :loading="deleteLoading"
              @click="openDelete(row.original)"
            />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Create / Edit Modal -->
    <UModal
      v-model:open="isModalOpen"
      :title="isEditing ? 'Edit Aplikasi' : 'Tambah Aplikasi Baru'"
    >
      <template #body>
        <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <UAlert
            v-if="formError"
            color="error"
            variant="soft"
            :description="formError"
            icon="i-lucide-circle-x"
          />

          <UFormField label="Nama Aplikasi" required>
            <UInput v-model="form.name" placeholder="Hi-DSign" class="w-full" />
          </UFormField>

          <UFormField label="Slug" required>
            <UInput v-model="form.slug" placeholder="hi-dsign" class="w-full" />
          </UFormField>

          <UFormField label="Base URL" required>
            <UInput v-model="form.base_url" placeholder="https://app.hrs-id.com" class="w-full" />
          </UFormField>

          <UFormField label="Icon URL">
            <UInput
              v-model="form.icon_url"
              placeholder="https://cdn.hrs-id.com/icon.png"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Deskripsi">
            <UTextarea
              v-model="form.description"
              placeholder="Deskripsi singkat aplikasi"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Akses Departemen">
            <div class="flex flex-col gap-2 mt-1">
              <label
                v-for="dept in departments"
                :key="dept.id"
                class="flex items-center gap-2 cursor-pointer select-none"
              >
                <UCheckbox
                  :model-value="form.department_ids.includes(dept.id)"
                  @update:model-value="toggleDept(dept.id)"
                />
                <span class="text-sm">{{ dept.name }}</span>
              </label>
            </div>
          </UFormField>

          <div class="flex justify-end gap-2 pt-2">
            <UButton label="Batal" color="neutral" variant="outline" @click="isModalOpen = false" />
            <UButton
              type="submit"
              :label="isEditing ? 'Simpan Perubahan' : 'Tambah'"
              :loading="formLoading"
            />
          </div>
        </form>
      </template>
    </UModal>

    <!-- Delete Confirmation Modal -->
    <UModal v-model:open="deleteModal.open" title="Hapus Aplikasi" :ui="{ footer: 'justify-end' }">
      <template #body>
        <div class="flex items-start gap-4">
          <div
            class="flex items-center justify-center rounded-full p-2 shrink-0 bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400"
          >
            <UIcon name="i-lucide-trash-2" class="w-5 h-5" />
          </div>
          <div>
            <p class="font-medium text-sm text-highlighted">Yakin ingin menghapus aplikasi ini?</p>
            <p class="text-sm text-muted mt-1">
              Aplikasi <span class="font-semibold text-default">{{ deleteModal.app?.name }}</span>
              akan dihapus beserta semua pengaturan aksesnya. Tindakan ini tidak bisa dibatalkan.
            </p>
          </div>
        </div>
      </template>
      <template #footer>
        <UButton
          label="Batal"
          color="neutral"
          variant="outline"
          @click="deleteModal.open = false"
        />
        <UButton label="Ya, Hapus" color="error" @click="handleDelete" />
      </template>
    </UModal>
  </div>
</template>
