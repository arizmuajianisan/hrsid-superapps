<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'admin'], layout: 'default' });

interface User {
  public_id: string;
  nik: string;
  email: string;
  full_name: string;
  role: string;
  department_id: number;
  is_active: boolean;
  created_at: string;
}

const { apiFetch } = useApi();
const toast = useToast();

// --- User list ---
const { data, refresh, status } = await useAsyncData('admin-users', () =>
  apiFetch<{ users: User[] }>('/api/v1/admin/users')
);
const users = computed(() => data.value?.users ?? []);

const columns = [
  { key: 'nik', label: 'NIK' },
  { key: 'full_name', label: 'Nama Lengkap' },
  { key: 'email', label: 'Email' },
  { key: 'role', label: 'Role' },
  { key: 'department_id', label: 'Dept.' },
  { key: 'is_active', label: 'Status' },
  { key: 'actions', label: 'Aksi' }
];

// --- Deactivate / Reactivate ---
const actionLoading = ref<string | null>(null);

async function deactivate(publicId: string) {
  actionLoading.value = publicId;
  try {
    await apiFetch(`/api/v1/admin/users/${publicId}/deactivate`, { method: 'POST' });
    toast.add({ title: 'Berhasil', description: 'User berhasil dinonaktifkan.', color: 'success' });
    await refresh();
  } catch {
    toast.add({ title: 'Gagal', description: 'Terjadi kesalahan. Silakan coba lagi.', color: 'error' });
  } finally {
    actionLoading.value = null;
  }
}

async function reactivate(publicId: string) {
  actionLoading.value = publicId;
  try {
    await apiFetch(`/api/v1/admin/users/${publicId}/reactivate`, { method: 'POST' });
    toast.add({ title: 'Berhasil', description: 'User berhasil diaktifkan.', color: 'success' });
    await refresh();
  } catch {
    toast.add({ title: 'Gagal', description: 'Terjadi kesalahan. Silakan coba lagi.', color: 'error' });
  } finally {
    actionLoading.value = null;
  }
}

// --- Create User modal ---
const isModalOpen = ref(false);
const createForm = reactive({
  nik: '',
  email: '',
  password: '',
  full_name: '',
  department_id: 1
});
const createError = ref<string | null>(null);
const createLoading = ref(false);

function openModal() {
  createError.value = null;
  Object.assign(createForm, { nik: '', email: '', password: '', full_name: '', department_id: 1 });
  isModalOpen.value = true;
}

async function handleCreateUser() {
  createError.value = null;
  createLoading.value = true;
  try {
    const config = useRuntimeConfig();
    await $fetch('/api/v1/register', {
      baseURL: config.public.apiBase,
      method: 'POST',
      body: {
        nik: createForm.nik,
        email: createForm.email,
        password: createForm.password,
        full_name: createForm.full_name,
        department_id: createForm.department_id
      }
    });
    isModalOpen.value = false;
    toast.add({ title: 'Berhasil', description: 'User baru berhasil dibuat.', color: 'success' });
    await refresh();
  } catch (err: unknown) {
    const fetchErr = err as { response?: { status?: number } };
    if (fetchErr.response?.status === 409) {
      createError.value = 'Email atau NIK sudah terdaftar.';
    } else {
      createError.value = 'Terjadi kesalahan. Silakan coba lagi.';
    }
  } finally {
    createLoading.value = false;
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold">
          Manajemen User
        </h1>
        <p class="text-muted mt-1">
          Kelola pengguna sistem HRSID
        </p>
      </div>
      <UButton
        label="Tambah User"
        icon="i-lucide-user-plus"
        @click="openModal"
      />
    </div>

    <UCard>
      <UTable
        :data="users"
        :columns="columns"
        :loading="status === 'pending'"
      >
        <template #role-cell="{ row }">
          <UBadge
            :color="row.original.role === 'admin' ? 'warning' : 'neutral'"
            variant="soft"
            :label="row.original.role"
          />
        </template>

        <template #is_active-cell="{ row }">
          <UBadge
            :color="row.original.is_active ? 'success' : 'error'"
            variant="soft"
            :label="row.original.is_active ? 'Aktif' : 'Nonaktif'"
          />
        </template>

        <template #actions-cell="{ row }">
          <UButton
            v-if="row.original.is_active"
            label="Nonaktifkan"
            color="error"
            variant="soft"
            size="xs"
            :loading="actionLoading === row.original.public_id"
            @click="deactivate(row.original.public_id)"
          />
          <UButton
            v-else
            label="Aktifkan"
            color="success"
            variant="soft"
            size="xs"
            :loading="actionLoading === row.original.public_id"
            @click="reactivate(row.original.public_id)"
          />
        </template>
      </UTable>
    </UCard>

    <!-- Create User Modal -->
    <UModal
      v-model:open="isModalOpen"
      title="Tambah User Baru"
    >
      <template #body>
        <form
          class="flex flex-col gap-4"
          @submit.prevent="handleCreateUser"
        >
          <UAlert
            v-if="createError"
            color="error"
            variant="soft"
            :description="createError"
            icon="i-lucide-circle-x"
          />

          <UFormField
            label="NIK"
            required
          >
            <UInput
              v-model="createForm.nik"
              placeholder="2024001"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Nama Lengkap"
            required
          >
            <UInput
              v-model="createForm.full_name"
              placeholder="John Doe"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Email"
            required
          >
            <UInput
              v-model="createForm.email"
              type="email"
              placeholder="john@hrs-id.com"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Password"
            required
          >
            <UInput
              v-model="createForm.password"
              type="password"
              placeholder="••••••••"
              class="w-full"
            />
          </UFormField>

          <UFormField
            label="Department ID"
            required
          >
            <UInput
              v-model.number="createForm.department_id"
              type="number"
              min="1"
              class="w-full"
            />
          </UFormField>

          <div class="flex justify-end gap-2 pt-2">
            <UButton
              label="Batal"
              color="neutral"
              variant="outline"
              @click="isModalOpen = false"
            />
            <UButton
              type="submit"
              label="Buat User"
              :loading="createLoading"
            />
          </div>
        </form>
      </template>
    </UModal>
  </div>
</template>
