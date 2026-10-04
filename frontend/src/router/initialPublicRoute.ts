export const initialPublicRouteLoaders = {
  '/login': () => import('@/views/auth/LoginView.vue'),
  '/register': () => import('@/views/auth/RegisterView.vue')
}

export async function preloadInitialPublicRoute(path: string): Promise<void> {
  if (path !== '/login' && path !== '/register') return
  const loader = initialPublicRouteLoaders[path]
  try {
    await loader()
  } catch {
    return
  }
}