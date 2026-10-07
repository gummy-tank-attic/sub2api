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
    class="flex min-h-screen flex-col bg-slate-50 text-slate-900"
  >
    <header class="border-b border-slate-200/80 px-4 py-4 sm:px-6 bg-white/80 backdrop-blur-md">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            :alt="siteName"
            class="h-[14px] w-auto max-w-[100px] shrink-0 object-contain"
          />
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-3">
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="text-xs font-medium text-slate-600 hover:text-slate-900 transition-colors"
          >
            Model Plaza
          </router-link>
          <LocaleSwitcher />
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-9 shrink-0 items-center justify-center rounded-lg bg-slate-900 px-4 py-1.5 text-xs font-semibold text-white hover:bg-slate-800 transition-colors"
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
            class="h-10 w-auto max-w-[220px] object-contain sm:h-12"
          />
        </div>
        <h1 class="sr-only">{{ siteName }}</h1>
        <h2 class="mt-2 text-2xl font-bold text-slate-900 sm:text-3xl sm:whitespace-nowrap tracking-normal leading-snug">{{ t('home.heroSubtitle') }}</h2>
        <p class="mt-3 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-slate-600 leading-relaxed">{{ t('home.heroDescription') }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-10 items-center justify-center rounded-xl bg-[#cc0000] px-6 py-2.5 text-sm font-semibold text-white shadow-md shadow-[#cc0000]/20 hover:bg-[#b30000] transition-all"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
        </router-link>
      </div>
    </main>

    <footer class="min-w-0 border-t border-slate-200/80 px-4 py-6 text-center text-xs text-slate-500 sm:px-6">
      <div>&copy; {{ currentYear }} {{ siteName }}. Operated by Helix Tech LLC. {{ t('home.footer.allRightsReserved') }}</div>
      <div class="mt-2 flex flex-wrap justify-center gap-4 text-xs text-slate-500">
        <router-link to="/legal/terms" class="transition-colors hover:text-slate-900">{{ t('home.termsOfService') }}</router-link>
        <span class="text-slate-300">&bull;</span>
        <router-link to="/legal/privacy" class="transition-colors hover:text-slate-900">{{ t('home.privacyPolicy') }}</router-link>
        <span class="text-slate-300">&bull;</span>
        <router-link to="/legal/refund" class="transition-colors hover:text-slate-900">{{ t('home.refundPolicy') }}</router-link>
        <template v-if="docUrl">
          <span class="text-slate-300">&bull;</span>
          <a :href="docUrl" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-slate-900">{{ t('home.docs') }}</a>
        </template>
      </div>
      <p class="mx-auto mt-3 max-w-2xl text-[11px] leading-relaxed text-slate-400">
        {{ t('home.disclaimer') }}
      </p>
    </footer>
  </div>

  <!-- Default Home Page: Pure Minimal Clean Light Mode -->
  <div
    v-else
    class="clean-canvas relative flex min-h-screen flex-col overflow-hidden text-slate-800 antialiased"
  >
    <!-- Header: Elevated Frosted Navbar -->
    <header class="sticky top-0 z-50 border-b border-slate-200/60 bg-white/80 px-6 py-3.5 backdrop-blur-md">
      <nav class="mx-auto flex max-w-7xl items-center justify-between">
        <!-- Left: Wordmark Logo & Status -->
        <div class="flex items-center gap-4 sm:gap-6">
          <router-link to="/" class="flex items-center transition-transform hover:scale-[1.01] duration-150">
            <img
              src="/logo.svg"
              :alt="siteName"
              class="h-[14px] w-auto max-w-[100px] object-contain"
            />
          </router-link>
          <div class="inline-flex items-center gap-1.5 rounded-full bg-slate-100/90 px-2.5 py-0.5 text-[11px] font-medium text-slate-600 border border-slate-200/60">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
            <span>{{ t('home.gatewayActive') }}</span>
          </div>
        </div>

        <!-- Right: Language & Action Controls -->
        <div class="flex items-center gap-3 sm:gap-4 text-sm">
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="text-xs sm:text-sm font-medium text-slate-600 hover:text-slate-900 transition-colors"
          >
            Model Plaza
          </router-link>
          <LocaleSwitcher />

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center justify-center rounded-lg bg-slate-900 px-3.5 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-slate-800 transition-all"
          >
            {{ t('home.dashboard') }} &rarr;
          </router-link>
          <div v-else class="flex items-center gap-3 sm:gap-4">
            <router-link
              to="/login"
              class="font-medium text-slate-600 hover:text-slate-900 transition-colors text-xs sm:text-sm"
            >
              {{ t('home.signIn') }}
            </router-link>
            <router-link
              to="/login"
              class="inline-flex items-center justify-center rounded-lg bg-slate-900 px-3.5 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-slate-800 transition-all"
            >
              {{ t('home.console') }} &rarr;
            </router-link>
          </div>
        </div>
      </nav>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 flex flex-col justify-center py-12 lg:py-16">
      <div class="mx-auto max-w-7xl px-6 lg:px-8 w-full space-y-14 lg:space-y-16">
        
        <!-- Hero Section: Split Grid Layout -->
        <div class="grid items-center gap-10 lg:grid-cols-12 lg:gap-14">
          
          <!-- Left Column: Copy & Actions -->
          <div class="lg:col-span-7 text-center lg:text-left space-y-6">
            
            <!-- Category Pill -->
            <div class="inline-flex items-center gap-2 rounded-full bg-rose-50/80 border border-rose-200/60 px-3 py-1 text-xs font-semibold text-[#cc0000]">
              <span>FXVIA</span>
              <span class="text-rose-300">•</span>
              <span class="text-slate-600 font-normal">AI API Gateway</span>
            </div>

            <!-- Main Value Headline: High Legibility, Comfortable Leading & Normal Tracking -->
            <h1 class="font-bold text-slate-900 text-3xl sm:text-4xl lg:text-[40px] leading-[1.35] tracking-normal">
              {{ t('home.heroSubtitle') }}
            </h1>
            
            <!-- Concise Value Subtitle: Relaxed Line Height -->
            <p class="text-base sm:text-lg text-slate-600 max-w-xl mx-auto lg:mx-0 font-normal leading-relaxed">
              {{ t('home.heroDescription') }}
            </p>

            <!-- Three Cohesive Feature Badges -->
            <div class="flex flex-wrap items-center justify-center lg:justify-start gap-2.5 pt-1 text-xs font-medium text-slate-600">
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 border border-slate-200/80 shadow-xs">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-400"></span>
                <span>{{ t('home.tags.subscriptionToApi') }}</span>
              </span>
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 border border-slate-200/80 shadow-xs">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-400"></span>
                <span>{{ t('home.tags.stickySession') }}</span>
              </span>
              <span class="inline-flex items-center gap-1.5 rounded-full bg-white px-3.5 py-1.5 border border-slate-200/80 shadow-xs">
                <span class="h-1.5 w-1.5 rounded-full bg-slate-400"></span>
                <span>{{ t('home.tags.realtimeBilling') }}</span>
              </span>
            </div>

            <!-- Primary Call to Action -->
            <div class="flex items-center justify-center lg:justify-start pt-3">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="inline-flex items-center justify-center rounded-xl bg-[#cc0000] px-6 py-3 text-sm font-semibold text-white shadow-md shadow-[#cc0000]/20 hover:bg-[#b30000] transition-all hover:scale-[1.01] active:scale-[0.98]"
              >
                <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
                <svg class="ml-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                </svg>
              </router-link>
            </div>
          </div>

          <!-- Right Column: Light Mode Developer Inspector Card -->
          <div class="lg:col-span-5 flex justify-center lg:justify-end">
            <div class="inspector-panel terminal-container w-full max-w-lg rounded-2xl overflow-hidden font-mono text-xs">
              <!-- Header Bar with Slate Mac Dots -->
              <div class="flex items-center justify-between border-b border-slate-200/80 bg-slate-50/80 px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="h-2.5 w-2.5 rounded-full bg-slate-300"></span>
                  <span class="h-2.5 w-2.5 rounded-full bg-slate-300"></span>
                  <span class="h-2.5 w-2.5 rounded-full bg-slate-300"></span>
                  <span class="ml-2 text-[11px] text-slate-500 font-sans font-medium">api.fxvia.com</span>
                </div>
                <div class="flex items-center gap-1.5 text-[10px] text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full font-mono font-medium">
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  <span>200 OK</span>
                </div>
              </div>

              <!-- Request Area -->
              <div class="p-5 space-y-3 leading-relaxed bg-white text-slate-700">
                <div class="flex items-start">
                  <span class="text-rose-600 mr-2 font-semibold">$</span>
                  <span class="text-slate-900 font-medium">curl -X POST https://api.fxvia.com/v1/chat/completions \</span>
                </div>
                <div class="pl-4 text-slate-600 space-y-1">
                  <div>-H <span class="text-slate-800 font-medium">"Authorization: Bearer sk-fxvia-***"</span> \</div>
                  <div>-H <span class="text-slate-800 font-medium">"Content-Type: application/json"</span> \</div>
                  <div>-d <span class="text-rose-700">'{"model": "claude-3-5-sonnet", ...}'</span></div>
                </div>

                <!-- Response Inset -->
                <div class="rounded-xl bg-slate-50 border border-slate-200/70 p-3.5 space-y-1 mt-2">
                  <div class="flex items-center justify-between text-[11px] text-emerald-700 font-semibold border-b border-slate-200/50 pb-1.5 mb-1.5">
                    <span>HTTP/2 200 OK</span>
                    <span class="text-[10px] text-slate-400 font-normal">latency: 124ms</span>
                  </div>
                  <div class="text-slate-700 leading-normal">
                    { "content": "Hello! Ready to assist." }
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Three Feature Value Cards Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4">
          <!-- Card 1: One API Key -->
          <div class="stripe-card rounded-2xl p-6 sm:p-7 flex flex-col justify-between">
            <div class="space-y-3.5">
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100/90 text-slate-700 border border-slate-200/80">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                  </svg>
                </div>
                <span class="text-[10px] font-mono tracking-widest uppercase text-slate-400 font-semibold">FEATURE 01</span>
              </div>
              <h3 class="text-lg font-bold text-slate-900 tracking-normal pt-1">{{ t('home.features.unifiedGateway') }}</h3>
              <p class="text-sm text-slate-600 leading-relaxed font-normal">
                {{ t('home.features.unifiedGatewayDesc') }}
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500 font-normal">
              <span>{{ t('home.features.unifiedGatewayTag') }}</span>
            </div>
          </div>

          <!-- Card 2: Smart Routing -->
          <div class="stripe-card rounded-2xl p-6 sm:p-7 flex flex-col justify-between">
            <div class="space-y-3.5">
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100/90 text-slate-700 border border-slate-200/80">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                </div>
                <span class="text-[10px] font-mono tracking-widest uppercase text-slate-400 font-semibold">FEATURE 02</span>
              </div>
              <h3 class="text-lg font-bold text-slate-900 tracking-normal pt-1">{{ t('home.features.multiAccount') }}</h3>
              <p class="text-sm text-slate-600 leading-relaxed font-normal">
                {{ t('home.features.multiAccountDesc') }}
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500 font-normal">
              <span>{{ t('home.features.multiAccountTag') }}</span>
            </div>
          </div>

          <!-- Card 3: Pay As You Go -->
          <div class="stripe-card rounded-2xl p-6 sm:p-7 flex flex-col justify-between">
            <div class="space-y-3.5">
              <div class="flex items-center justify-between">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100/90 text-slate-700 border border-slate-200/80">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <span class="text-[10px] font-mono tracking-widest uppercase text-slate-400 font-semibold">FEATURE 03</span>
              </div>
              <h3 class="text-lg font-bold text-slate-900 tracking-normal pt-1">{{ t('home.features.balanceQuota') }}</h3>
              <p class="text-sm text-slate-600 leading-relaxed font-normal">
                {{ t('home.features.balanceQuotaDesc') }}
              </p>
            </div>
            <div class="mt-6 pt-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500 font-normal">
              <span>{{ t('home.features.balanceQuotaTag') }}</span>
            </div>
          </div>
        </div>

      </div>
    </main>

    <!-- Footer: Clean Legal & Attribution -->
    <footer class="border-t border-slate-200/80 bg-white/60 px-6 py-8 text-xs text-slate-500 backdrop-blur-xs">
      <div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 sm:flex-row">
        <div>
          &copy; {{ currentYear }} {{ siteName }}. Operated by Helix Tech LLC. {{ t('home.footer.allRightsReserved') }}
        </div>
        <div class="flex flex-wrap items-center justify-center gap-6">
          <router-link to="/legal/terms" class="transition-colors hover:text-slate-900">
            {{ t('home.termsOfService') }}
          </router-link>
          <span class="text-slate-300">&bull;</span>
          <router-link to="/legal/privacy" class="transition-colors hover:text-slate-900">
            {{ t('home.privacyPolicy') }}
          </router-link>
          <span class="text-slate-300">&bull;</span>
          <router-link to="/legal/refund" class="transition-colors hover:text-slate-900">
            {{ t('home.refundPolicy') }}
          </router-link>
          <template v-if="docUrl">
            <span class="text-slate-300">&bull;</span>
            <a :href="docUrl" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-slate-900">
              {{ t('home.docs') }}
            </a>
          </template>
        </div>
      </div>
      <div class="mx-auto mt-4 max-w-7xl border-t border-slate-200/50 pt-3">
        <p class="text-[11px] leading-relaxed text-slate-400">
          {{ t('home.disclaimer') }}
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import DOMPurify from 'dompurify'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'FXVIA')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
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

// Model Plaza entry
const modelPlazaEnabled = computed(() => Boolean(appStore.cachedPublicSettings?.model_plaza_enabled))
const modelPlazaRequiresAuth = computed(() => appStore.cachedPublicSettings?.model_plaza_require_auth === true)
const showModelPlazaEntry = computed(() => modelPlazaEnabled.value && (isAuthenticated.value || !modelPlazaRequiresAuth.value))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  // Always enforce clean light mode
  document.documentElement.classList.remove('dark')
  localStorage.setItem('theme', 'light')
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style>
/* Clean soothing alabaster canvas without distracting grid lines */
.clean-canvas {
  background-color: #f8fafc;
  background-image: radial-gradient(ellipse at 50% -20%, rgba(204, 0, 0, 0.035) 0%, transparent 65%);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans SC", sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

.stripe-card {
  background: #ffffff;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 
    0 1px 3px 0 rgba(15, 23, 42, 0.02),
    0 10px 24px -4px rgba(15, 23, 42, 0.03);
  transition: all 0.2s ease;
}

.stripe-card:hover {
  border-color: rgba(203, 213, 225, 1);
  box-shadow: 
    0 4px 6px -1px rgba(15, 23, 42, 0.03),
    0 16px 30px -6px rgba(15, 23, 42, 0.04);
  transform: translateY(-2px);
}

.inspector-panel {
  background: #ffffff;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 
    0 4px 6px -1px rgba(15, 23, 42, 0.02),
    0 20px 40px -8px rgba(15, 23, 42, 0.04);
}
</style>
