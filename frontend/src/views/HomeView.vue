<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="homeContentUrl"
      :src="homeContentUrl"
      class="h-screen w-full border-0"
      sandbox="allow-scripts allow-forms allow-popups"
      referrerpolicy="no-referrer"
      allowfullscreen
    ></iframe>
    <div v-else v-html="sanitizedHomeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-slate-50 text-slate-900 dark:bg-dark-950 dark:text-white"
  >
    <header class="border-b border-slate-200/80 px-4 py-4 sm:px-6 dark:border-dark-800">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            :alt="siteName"
            class="h-7 w-auto max-w-[130px] shrink-0 object-contain"
          />
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <LocaleSwitcher />
          <button
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-slate-500 hover:bg-slate-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 shrink-0 items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <div class="mx-auto mb-6 flex justify-center">
          <img
            src="/logo.svg"
            :alt="siteName"
            class="h-16 w-auto max-w-[300px] object-contain sm:h-20 sm:max-w-[380px]"
          />
        </div>
        <h1 class="sr-only">{{ siteName }}</h1>
        <h2 class="mt-2 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl sm:whitespace-nowrap">One API. Multiple AI Models.</h2>
        <p class="mt-3 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-slate-600 dark:text-dark-300">Unified access, smart routing, and usage-based billing.</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-10 items-center justify-center rounded-xl bg-[#be123c] px-6 py-2.5 text-sm font-semibold text-white shadow-md shadow-[#be123c]/20 hover:bg-[#9f1239] transition-all"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : 'Get Started' }}
        </router-link>
      </div>
    </main>

    <footer class="min-w-0 border-t border-slate-200/80 px-4 py-6 text-center text-xs text-slate-500 sm:px-6 dark:border-dark-800 dark:text-dark-400">
      <div>&copy; {{ currentYear }} {{ siteName }}. Operated by Helix Tech LLC. All rights reserved.</div>
      <div class="mt-2 flex flex-wrap justify-center gap-4 text-xs text-slate-500 dark:text-dark-400">
        <router-link to="/legal/terms" class="transition-colors hover:text-slate-900 dark:hover:text-white">Terms of Service</router-link>
        <span class="text-slate-300 dark:text-dark-700">&bull;</span>
        <router-link to="/legal/privacy" class="transition-colors hover:text-slate-900 dark:hover:text-white">Privacy Policy</router-link>
        <span class="text-slate-300 dark:text-dark-700">&bull;</span>
        <router-link to="/legal/refund" class="transition-colors hover:text-slate-900 dark:hover:text-white">Refund Policy</router-link>
      </div>
    </footer>
  </div>

  <!-- Default Home Page (Option A: Stripe Editorial Warm Alabaster + Ruby Accent) -->
  <div
    v-else
    class="stripe-canvas relative flex min-h-screen flex-col overflow-hidden text-slate-800 antialiased dark:bg-dark-950 dark:text-white"
  >
    <!-- Header: Elevated Frosted Navbar -->
    <header class="sticky top-0 z-50 border-b border-slate-200/60 bg-white/80 px-6 py-3.5 backdrop-blur-md dark:border-dark-800/60 dark:bg-dark-900/80">
      <nav class="mx-auto flex max-w-7xl items-center justify-between">
        <!-- Left: Wordmark Logo & Status -->
        <div class="flex items-center gap-6">
          <router-link to="/" class="flex items-center gap-2 transition-transform hover:scale-[1.01] duration-150">
            <img
              src="/logo.svg"
              :alt="siteName"
              class="h-7 w-auto max-w-[160px] object-contain"
            />
          </router-link>
          <div class="hidden sm:inline-flex items-center gap-1.5 rounded-full bg-slate-100/80 px-2.5 py-0.5 text-[11px] font-medium text-slate-600 border border-slate-200/50 dark:bg-dark-800 dark:text-dark-300 dark:border-dark-700">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
            <span>Gateway Active</span>
          </div>
        </div>

        <!-- Right: Actions -->
        <div class="flex items-center gap-4 text-sm">
          <LocaleSwitcher />
          
          <button
            @click="toggleTheme"
            class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white shadow-sm hover:bg-slate-800 transition-all dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
          >
            {{ t('home.dashboard') }} &rarr;
          </router-link>
          <div v-else class="flex items-center gap-4">
            <router-link
              to="/login"
              class="font-medium text-slate-600 hover:text-slate-900 transition-colors dark:text-dark-300 dark:hover:text-white"
            >
              Sign In
            </router-link>
            <router-link
              to="/login"
              class="inline-flex items-center justify-center rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white shadow-sm hover:bg-slate-800 transition-all dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
            >
              Console &rarr;
            </router-link>
          </div>
        </div>
      </nav>
    </header>

    <!-- Main Content Area -->
    <main class="relative z-10 flex-1 flex flex-col justify-center px-6 py-10 lg:py-16">
      <div class="mx-auto max-w-7xl w-full space-y-12 lg:space-y-14">
        
        <!-- Hero Section - Left/Right Layout -->
        <div class="grid items-center gap-10 lg:grid-cols-12 lg:gap-14">
          <!-- Left: Text Content -->
          <div class="lg:col-span-7 text-center lg:text-left">
            <div class="mb-5 flex justify-center lg:justify-start">
              <img
                src="/logo.svg"
                :alt="siteName"
                class="h-10 w-auto max-w-[240px] object-contain sm:h-12 sm:max-w-[280px] drop-shadow-xs transition-transform hover:scale-105 duration-300"
              />
            </div>

            <h1
              class="heading-font mb-4 text-2xl sm:text-3xl lg:text-[37px] xl:text-[41px] font-bold tracking-tight text-slate-900 dark:text-white lg:whitespace-nowrap leading-[1.14]"
            >
              One API. Multiple AI Models.
            </h1>
            <p class="mb-6 text-base text-slate-600 dark:text-dark-300 md:text-lg leading-relaxed max-w-xl mx-auto lg:mx-0">
              Unified access, smart routing, and usage-based billing.
            </p>

            <!-- Feature Tags -->
            <div class="mb-8 flex flex-wrap items-center justify-center lg:justify-start gap-2.5 text-xs font-medium text-slate-700 dark:text-dark-200">
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 shadow-sm border border-slate-200/80 dark:bg-dark-900 dark:border-dark-700">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-800 dark:bg-slate-200"></span>
                Unified API
              </span>
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 shadow-sm border border-slate-200/80 dark:bg-dark-900 dark:border-dark-700">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-800 dark:bg-slate-200"></span>
                Smart Routing
              </span>
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 shadow-sm border border-slate-200/80 dark:bg-dark-900 dark:border-dark-700">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-800 dark:bg-slate-200"></span>
                Pay As You Go
              </span>
            </div>

            <!-- CTA Button (Deep Ruby Carmine) -->
            <div class="flex flex-col sm:flex-row items-center justify-center lg:justify-start gap-4">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="inline-flex items-center justify-center rounded-xl bg-[#be123c] hover:bg-[#9f1239] px-7 py-3 text-sm font-semibold text-white shadow-md shadow-[#be123c]/20 transition-all active:scale-[0.98]"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : 'Get Started' }}
                <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
              </router-link>
            </div>
          </div>

          <!-- Right: Light Mode Developer Inspector Panel (Seamless, No Black Hole) -->
          <div class="lg:col-span-5 flex justify-center lg:justify-end">
            <div class="inspector-panel w-full max-w-[460px] rounded-2xl overflow-hidden font-mono text-xs">
              <!-- Window Header -->
              <div class="flex items-center justify-between border-b border-slate-100 bg-slate-50/80 px-4 py-3 dark:border-dark-800 dark:bg-dark-900">
                <div class="flex items-center gap-2">
                  <div class="h-2.5 w-2.5 rounded-full bg-slate-300 dark:bg-dark-700"></div>
                  <div class="h-2.5 w-2.5 rounded-full bg-slate-300 dark:bg-dark-700"></div>
                  <div class="h-2.5 w-2.5 rounded-full bg-slate-300 dark:bg-dark-700"></div>
                  <span class="ml-2 text-[11px] text-slate-500 font-sans font-medium dark:text-dark-400">api.fxvia.com</span>
                </div>
                <div class="flex items-center gap-1.5 text-[10px] text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full font-mono dark:bg-emerald-950/60 dark:text-emerald-400 dark:border-emerald-800/60">
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  <span>200 OK</span>
                </div>
              </div>

              <!-- Request Area -->
              <div class="p-5 space-y-3 leading-relaxed bg-white dark:bg-dark-950 text-slate-700 dark:text-dark-200">
                <div class="flex items-start">
                  <span class="text-rose-600 mr-2 font-semibold">$</span>
                  <span class="text-slate-900 dark:text-white font-medium">curl -X POST https://api.fxvia.com/v1/chat/completions \</span>
                </div>
                <div class="pl-4 text-slate-600 dark:text-dark-400 space-y-1">
                  <div>-H <span class="text-slate-800 dark:text-dark-200 font-medium">"Authorization: Bearer sk-fxvia-***"</span> \</div>
                  <div>-H <span class="text-slate-800 dark:text-dark-200 font-medium">"Content-Type: application/json"</span> \</div>
                  <div>-d <span class="text-rose-700 dark:text-rose-400">'{"model": "claude-3-5-sonnet", ...}'</span></div>
                </div>

                <!-- Response Inset -->
                <div class="rounded-xl bg-slate-50 border border-slate-200/70 p-3.5 space-y-1 mt-2 dark:bg-dark-900 dark:border-dark-800">
                  <div class="flex items-center justify-between text-[11px]">
                    <span class="text-emerald-700 dark:text-emerald-400 font-semibold font-mono">HTTP/2 200 OK</span>
                    <span class="text-slate-400 dark:text-dark-500 font-mono text-[10px]">latency: 124ms</span>
                  </div>
                  <div class="text-[11px] text-slate-700 dark:text-dark-300 font-mono">
                    { "content": "Hello! Ready to assist." }
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Three Core Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8 pt-4">
          <!-- Card 1: One API Key -->
          <div class="stripe-card rounded-2xl p-7 flex flex-col justify-between dark:bg-dark-900 dark:border-dark-800">
            <div>
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-50 border border-slate-200 text-slate-700 dark:bg-dark-800 dark:border-dark-700 dark:text-dark-200">
                  <Icon name="key" size="md" />
                </div>
                <span class="text-[10px] font-mono text-slate-400 font-medium tracking-wider uppercase dark:text-dark-500">FEATURE 01</span>
              </div>
              <h3 class="mt-5 text-lg font-bold text-slate-900 dark:text-white tracking-tight heading-font">
                One API Key
              </h3>
              <p class="mt-2 text-sm text-slate-600 dark:text-dark-300 leading-relaxed">
                Access supported AI models with a single key.
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 dark:border-dark-800 text-xs text-slate-400 dark:text-dark-500 font-medium flex items-center gap-2">
              <span class="h-1.5 w-1.5 rounded-full bg-slate-300 dark:bg-dark-600"></span>
              <span>Single integration &bull; Unified token</span>
            </div>
          </div>

          <!-- Card 2: Smart Routing -->
          <div class="stripe-card rounded-2xl p-7 flex flex-col justify-between dark:bg-dark-900 dark:border-dark-800">
            <div>
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-50 border border-slate-200 text-slate-700 dark:bg-dark-800 dark:border-dark-700 dark:text-dark-200">
                  <Icon name="swap" size="md" />
                </div>
                <span class="text-[10px] font-mono text-slate-400 font-medium tracking-wider uppercase dark:text-dark-500">FEATURE 02</span>
              </div>
              <h3 class="mt-5 text-lg font-bold text-slate-900 dark:text-white tracking-tight heading-font">
                Smart Routing
              </h3>
              <p class="mt-2 text-sm text-slate-600 dark:text-dark-300 leading-relaxed">
                Automatic routing and failover across upstream providers.
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 dark:border-dark-800 text-xs text-slate-400 dark:text-dark-500 font-medium flex items-center gap-2">
              <span class="h-1.5 w-1.5 rounded-full bg-slate-300 dark:bg-dark-600"></span>
              <span>High availability &bull; Multi-region redundancy</span>
            </div>
          </div>

          <!-- Card 3: Pay As You Go -->
          <div class="stripe-card rounded-2xl p-7 flex flex-col justify-between dark:bg-dark-900 dark:border-dark-800">
            <div>
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-50 border border-slate-200 text-slate-700 dark:bg-dark-800 dark:border-dark-700 dark:text-dark-200">
                  <Icon name="chart" size="md" />
                </div>
                <span class="text-[10px] font-mono text-slate-400 font-medium tracking-wider uppercase dark:text-dark-500">FEATURE 03</span>
              </div>
              <h3 class="mt-5 text-lg font-bold text-slate-900 dark:text-white tracking-tight heading-font">
                Pay As You Go
              </h3>
              <p class="mt-2 text-sm text-slate-600 dark:text-dark-300 leading-relaxed">
                Pay for what you use. Track usage and costs.
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 dark:border-dark-800 text-xs text-slate-400 dark:text-dark-500 font-medium flex items-center gap-2">
              <span class="h-1.5 w-1.5 rounded-full bg-slate-300 dark:bg-dark-600"></span>
              <span>Transparent metrics &bull; Real-time metering</span>
            </div>
          </div>
        </div>

      </div>
    </main>

    <!-- Footer -->
    <footer class="relative z-10 border-t border-slate-200/60 bg-white/70 px-6 py-6 text-xs text-slate-500 backdrop-blur-sm dark:border-dark-800/60 dark:bg-dark-900/70 dark:text-dark-400">
      <div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 text-center sm:flex-row sm:text-left">
        <div>
          &copy; {{ currentYear }} {{ siteName }}. Operated by Helix Tech LLC. All rights reserved.
        </div>
        <div class="flex flex-wrap items-center justify-center gap-6">
          <router-link to="/legal/terms" class="transition-colors hover:text-slate-900 dark:hover:text-white">
            Terms of Service
          </router-link>
          <span class="text-slate-300 dark:text-dark-700">&bull;</span>
          <router-link to="/legal/privacy" class="transition-colors hover:text-slate-900 dark:hover:text-white">
            Privacy Policy
          </router-link>
          <span class="text-slate-300 dark:text-dark-700">&bull;</span>
          <router-link to="/legal/refund" class="transition-colors hover:text-slate-900 dark:hover:text-white">
            Refund Policy
          </router-link>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import DOMPurify from 'dompurify'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'FXVIA')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const homeContentUrl = computed(() => {
  return /^https?:\/\//i.test(homeContent.value.trim()) ? sanitizeUrl(homeContent.value) : ''
})
const sanitizedHomeContent = computed(() => DOMPurify.sanitize(homeContent.value, {
  USE_PROFILES: { html: true },
  FORBID_TAGS: ['form', 'input', 'button', 'textarea', 'select', 'style', 'iframe'],
  FORBID_ATTR: ['style'],
}))
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    isDark.value = true
    document.documentElement.classList.add('dark')
  } else {
    isDark.value = false
    document.documentElement.classList.remove('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.heading-font {
  font-family: 'Plus Jakarta Sans', 'Inter', -apple-system, sans-serif;
}
</style>

<style>
.stripe-canvas {
  background-color: #f8fafc;
  background-image: 
    radial-gradient(ellipse at 50% -20%, rgba(190, 18, 60, 0.035) 0%, transparent 60%),
    linear-gradient(rgba(15, 23, 42, 0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.02) 1px, transparent 1px);
  background-size: 100% 100%, 40px 40px, 40px 40px;
}

.dark .stripe-canvas {
  background-color: #07090e !important;
  background-image: 
    radial-gradient(ellipse at 50% -20%, rgba(225, 29, 72, 0.12) 0%, transparent 60%),
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px) !important;
}

.stripe-card {
  background: #ffffff;
  border: 1px solid rgba(203, 213, 225, 0.6);
  box-shadow: 
    0 1px 3px 0 rgba(15, 23, 42, 0.02),
    0 10px 24px -4px rgba(15, 23, 42, 0.03);
  transition: all 0.25s ease;
}

.stripe-card:hover {
  border-color: rgba(148, 163, 184, 0.8);
  box-shadow: 
    0 2px 4px 0 rgba(15, 23, 42, 0.02),
    0 16px 32px -4px rgba(15, 23, 42, 0.05);
  transform: translateY(-2px);
}

.dark .stripe-card {
  background: rgba(15, 23, 42, 0.6) !important;
  border-color: rgba(255, 255, 255, 0.08) !important;
  box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.5) !important;
}

.inspector-panel {
  background: #ffffff;
  border: 1px solid rgba(203, 213, 225, 0.7);
  box-shadow: 
    0 4px 6px -1px rgba(15, 23, 42, 0.02),
    0 20px 40px -8px rgba(15, 23, 42, 0.05);
}

.dark .inspector-panel {
  background: #090d16 !important;
  border-color: rgba(255, 255, 255, 0.12) !important;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.7) !important;
}
</style>
