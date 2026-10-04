import { afterEach, describe, expect, it, vi } from 'vitest'
import { initialPublicRouteLoaders, preloadInitialPublicRoute } from '../initialPublicRoute'

afterEach(() => vi.restoreAllMocks())

describe('preloadInitialPublicRoute', () => {
  it.each(['/login', '/register'] as const)('loads only the matching public route %s', async (path) => {
    const login = vi.spyOn(initialPublicRouteLoaders, '/login').mockResolvedValue({ default: {} } as never)
    const register = vi.spyOn(initialPublicRouteLoaders, '/register').mockResolvedValue({ default: {} } as never)
    await preloadInitialPublicRoute(path)
    expect(login).toHaveBeenCalledTimes(path === '/login' ? 1 : 0)
    expect(register).toHaveBeenCalledTimes(path === '/register' ? 1 : 0)
  })

  it.each(['/admin/dashboard', '/dashboard', '/auth/callback', '/login/unknown', '/', 'constructor', '__proto__'])('does not preload other routes %s', async (path) => {
    const login = vi.spyOn(initialPublicRouteLoaders, '/login')
    const register = vi.spyOn(initialPublicRouteLoaders, '/register')
    await preloadInitialPublicRoute(path)
    expect(login).not.toHaveBeenCalled()
    expect(register).not.toHaveBeenCalled()
  })

  it('allows navigation to retry after speculative loading fails', async () => {
    const loader = vi.spyOn(initialPublicRouteLoaders, '/login')
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ default: {} } as never)
    await expect(preloadInitialPublicRoute('/login')).resolves.toBeUndefined()
    await expect(initialPublicRouteLoaders['/login']()).resolves.toEqual({ default: {} })
    expect(loader).toHaveBeenCalledTimes(2)
  })
})