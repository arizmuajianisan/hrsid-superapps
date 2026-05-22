<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 p-4">
    <div class="w-full max-w-md">
      <div class="flex flex-col items-center mb-8 gap-2">
        <UIcon name="i-lucide-shield-check" class="size-12 text-primary" />
        <h1 class="text-2xl font-bold">HRSID</h1>
        <p class="text-sm text-muted">Hirose Identity & Access Management</p>
      </div>
      <slot />
    </div>
  </div>
</template>
