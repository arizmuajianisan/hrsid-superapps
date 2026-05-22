export default defineNuxtPlugin(async () => {
  const auth = useAuthStore()
  if (!auth.isAuthenticated) {
    await auth.initAuth()
  }
})
