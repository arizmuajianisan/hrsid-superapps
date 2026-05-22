<script setup lang="ts">
definePageMeta({ layout: "auth" });

const auth = useAuthStore();
const router = useRouter();

if (auth.isAuthenticated) {
  await navigateTo("/");
}

const form = reactive({ identifier: "", password: "" });
const errorMessage = ref<string | null>(null);
const isLoading = ref(false);

async function handleLogin() {
  errorMessage.value = null;
  isLoading.value = true;
  try {
    await auth.login(form.identifier, form.password);
    await router.push("/");
  } catch {
    errorMessage.value = "Identifier atau password salah. Silakan coba lagi.";
  } finally {
    isLoading.value = false;
  }
}
</script>

<template>
  <UCard>
    <template #header>
      <h2 class="text-xl font-semibold text-center">
        Masuk ke Akun Anda
      </h2>
    </template>

    <form
      class="flex flex-col gap-4"
      @submit.prevent="handleLogin"
    >
      <UAlert
        v-if="errorMessage"
        color="error"
        variant="soft"
        :description="errorMessage"
        icon="i-lucide-circle-x"
      />

      <UFormField
        label="Identifier (Email / NIK)"
        required
      >
        <UInput
          v-model="form.identifier"
          placeholder="email@hrs-id.com atau NIK"
          autocomplete="username"
          class="w-full"
        />
      </UFormField>

      <UFormField
        label="Password"
        required
      >
        <UInput
          v-model="form.password"
          type="password"
          placeholder="••••••••"
          autocomplete="current-password"
          class="w-full"
        />
      </UFormField>

      <UButton
        type="submit"
        label="Masuk"
        block
        :loading="isLoading"
        class="mt-2"
      />
    </form>
  </UCard>
</template>
