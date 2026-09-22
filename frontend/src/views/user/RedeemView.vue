<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6 pb-12">
      <!-- 顶部账户与汇率信息横幅 (匹配站点 Primary 青绿色系) -->
      <div class="relative overflow-hidden rounded-2xl border border-primary-100 bg-gradient-to-br from-primary-50/60 via-teal-50/30 to-white p-6 sm:p-8 shadow-sm">
        <div class="relative z-10 flex flex-col md:flex-row md:items-center md:justify-between gap-6">
          <div class="space-y-2">
            <div class="inline-flex items-center gap-2 rounded-full border border-primary-200/80 bg-white px-3 py-1 text-xs font-semibold text-primary-700 shadow-sm">
              <span class="inline-block h-2 w-2 rounded-full bg-primary-500 animate-pulse"></span>
              云猫官方通道 · 自动秒发
            </div>
            <h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-gray-900">
              CNY 充值 · 卡密兑换
            </h1>
            <p class="text-sm text-gray-600 max-w-xl leading-relaxed">
              支持微信与支付宝付款购买充值卡，或直接输入已有卡密。兑换后秒级充入账户余额，永不过期。
            </p>
          </div>

          <!-- 右侧账户与汇率卡片 (移除单独的并发大块，融入余额信息) -->
          <div class="flex flex-wrap sm:flex-nowrap items-center gap-3">
            <div class="flex-1 min-w-[150px] rounded-xl border border-primary-100 bg-white p-4 shadow-sm">
              <div class="text-xs font-medium text-gray-500">{{ t('redeem.currentBalance') }}</div>
              <div class="mt-1 text-2xl font-bold text-gray-900">
                ${{ user?.balance?.toFixed(2) || '0.00' }}
              </div>
              <div class="mt-1 text-xs text-gray-500 flex items-center gap-1.5">
                <span class="inline-block h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                <span>{{ t('redeem.concurrency') }}: {{ user?.concurrency || 0 }} {{ t('redeem.requests') }}</span>
              </div>
            </div>
            <div class="flex-1 min-w-[140px] rounded-xl border border-amber-200/80 bg-white p-4 shadow-sm">
              <div class="text-xs font-medium text-gray-500">充值换算比率</div>
              <div class="mt-1 text-2xl font-bold text-amber-600">
                1 : 1
              </div>
              <div class="mt-1 text-xs text-gray-500">￥1.00 = $1.00 额度</div>
            </div>
          </div>
        </div>

        <div class="pointer-events-none absolute -right-12 -bottom-12 h-64 w-64 rounded-full bg-primary-200/20 blur-3xl"></div>
      </div>

      <!-- 模块 1：挑选充值卡面额 (CNY充值货架) -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-primary-600 text-white shadow-sm font-semibold text-sm">
              1
            </div>
            <h2 class="text-lg font-bold text-gray-900">挑选充值卡面额</h2>
            <span class="text-xs text-gray-500">（点击直接直达云猫店铺安全购买）</span>
          </div>
          <a
            :href="yunmaoShopUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center text-xs font-medium text-primary-600 hover:text-primary-700 hover:underline gap-1"
          >
            直接访问云猫小铺
            <Icon name="externalLink" size="xs" />
          </a>
        </div>

        <!-- 5 个极简充值卡卡片 -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
          <div
            v-for="card in cards"
            :key="card.amount"
            class="group relative flex flex-col justify-between rounded-2xl border border-gray-200/90 bg-white p-4 sm:p-5 shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-primary-400 hover:shadow-md text-center"
          >
            <div>
              <div class="flex items-baseline justify-center gap-0.5">
                <span class="text-3xl font-extrabold text-gray-900">${{ card.amount }}</span>
                <span class="text-xs font-medium text-gray-500">额度</span>
              </div>

              <div class="mt-3 rounded-xl bg-gray-50 border border-gray-100 py-2 px-2.5 text-center">
                <span class="text-xs text-gray-500">售价：</span>
                <span class="text-base font-bold text-primary-600">¥{{ card.price }}</span>
              </div>
            </div>

            <div class="mt-4 pt-3 border-t border-gray-100">
              <button
                type="button"
                @click="openShop(card)"
                class="w-full inline-flex items-center justify-center gap-1.5 rounded-xl bg-gray-900 hover:bg-primary-600 text-white py-2 px-2 text-xs font-semibold transition-colors duration-150 shadow-sm"
              >
                <span>立即购买</span>
                <Icon name="externalLink" size="xs" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 模块 2：原地秒充 · 兑换卡密 -->
      <div id="redeem-section" class="rounded-2xl border border-primary-200/80 bg-gradient-to-br from-primary-50/30 via-white to-white p-6 shadow-sm">
        <div class="flex items-center gap-2 mb-4">
          <div class="flex h-7 w-7 items-center justify-center rounded-lg bg-primary-600 text-white shadow-sm font-semibold text-sm">
            2
          </div>
          <div>
            <h2 class="text-lg font-bold text-gray-900">原地秒充 · 兑换卡密立即到账</h2>
            <p class="text-xs text-gray-500">在云猫付款成功后复制得到的卡密，在此处粘贴即可，无需跳转其他页面</p>
          </div>
        </div>

        <form @submit.prevent="handleRedeem" class="mt-4">
          <div class="flex flex-col sm:flex-row gap-3">
            <div class="relative flex-1">
              <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400">
                <Icon name="gift" size="md" />
              </div>
              <input
                id="code"
                v-model="redeemCode"
                type="text"
                required
                :placeholder="t('redeem.redeemCodePlaceholder')"
                :disabled="submitting"
                class="w-full rounded-xl border border-gray-300 bg-white py-3 pl-11 pr-4 text-sm font-medium text-gray-900 placeholder:text-gray-400 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100 transition-all"
              />
            </div>

            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="btn btn-primary px-6 py-3 text-sm font-semibold"
            >
              <svg
                v-if="submitting"
                class="h-4 w-4 animate-spin text-white"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <Icon v-else name="checkCircle" size="sm" />
              <span>{{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}</span>
            </button>
          </div>
          <p class="mt-2 text-xs text-gray-500">
            {{ t('redeem.redeemCodeHint') }}
          </p>
        </form>

        <!-- 兑换成功提示 (完整保留余额、并发、订阅等不同类型的反馈) -->
        <transition name="fade">
          <div
            v-if="redeemResult"
            class="mt-4 rounded-xl border border-primary-200 bg-primary-50/60 p-4"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-100 text-primary-600">
                <Icon name="checkCircle" size="md" />
              </div>
              <div class="flex-1">
                <h4 class="text-sm font-bold text-primary-900">
                  {{ t('redeem.redeemSuccess') }}
                </h4>
                <p class="mt-1 text-xs text-primary-700">
                  {{ redeemResult.message }}
                </p>
                <div class="mt-2 space-y-1 text-xs text-primary-800">
                  <p v-if="redeemResult.type === 'balance'" class="font-medium">
                    {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                  </p>
                  <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                    {{ t('redeem.added') }}: {{ redeemResult.value }}
                    {{ t('redeem.concurrentRequests') }}
                  </p>
                  <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                    {{ t('redeem.subscriptionAssigned') }}
                    <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                    <span v-if="redeemResult.validity_days">
                      ({{ t('redeem.subscriptionDays', { days: redeemResult.validity_days }) }})
                    </span>
                  </p>
                  <p v-if="redeemResult.new_balance !== undefined">
                    {{ t('redeem.newBalance') }}:
                    <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                  </p>
                  <p v-if="redeemResult.new_concurrency !== undefined">
                    {{ t('redeem.newConcurrency') }}:
                    <span class="font-semibold">{{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}</span>
                  </p>
                </div>
              </div>
            </div>
          </div>
        </transition>

        <!-- 错误提示 -->
        <transition name="fade">
          <div
            v-if="errorMessage"
            class="mt-4 rounded-xl border border-red-200 bg-red-50 p-4"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-red-100 text-red-600">
                <Icon name="exclamationCircle" size="md" />
              </div>
              <div class="flex-1">
                <h4 class="text-sm font-bold text-red-900">{{ t('redeem.redeemFailed') }}</h4>
                <p class="mt-1 text-xs text-red-700">{{ errorMessage }}</p>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- 模块 3：关于兑换码须知 (原兑换页规则，100% 完整保留) -->
      <div class="rounded-2xl border border-primary-100 bg-primary-50/30 p-6 shadow-sm">
        <div class="flex items-start gap-4">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-100 text-primary-600">
            <Icon name="infoCircle" size="md" />
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-bold text-primary-900">
              {{ t('redeem.aboutCodes') }}
            </h3>
            <ul class="mt-2 list-inside list-disc space-y-1 text-xs text-primary-800">
              <li>{{ t('redeem.codeRule1') }}</li>
              <li>{{ t('redeem.codeRule2') }}</li>
              <li>
                {{ t('redeem.codeRule3') }}
                <span
                  v-if="contactInfo"
                  class="ml-1.5 inline-flex items-center rounded-md bg-primary-200/60 px-2 py-0.5 text-xs font-medium text-primary-900"
                >
                  {{ contactInfo }}
                </span>
              </li>
              <li>{{ t('redeem.codeRule4') }}</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 模块 4：最近兑换记录 (原兑换页功能，100% 完整保留) -->
      <div class="rounded-2xl border border-gray-200/80 bg-white shadow-sm overflow-hidden">
        <div class="border-b border-gray-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-base font-bold text-gray-900">
            {{ t('redeem.recentActivity') }}
          </h2>
          <span class="text-xs text-gray-500">{{ t('common.total') }}: {{ historyTotal }} {{ t('pagination.results') }}</span>
        </div>
        <div class="p-6">
          <!-- 加载状态 -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-600" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <!-- 历史列表 -->
          <div v-else-if="history.length > 0" class="space-y-3">
            <div
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between rounded-xl bg-gray-50/80 border border-gray-100 p-4"
            >
              <div class="flex items-center gap-4">
                <div
                  :class="[
                    'flex h-10 w-10 items-center justify-center rounded-xl',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'bg-primary-100 text-primary-600'
                        : 'bg-red-100 text-red-600'
                      : isSubscriptionType(item.type)
                        ? 'bg-purple-100 text-purple-600'
                        : item.value >= 0
                          ? 'bg-teal-100 text-teal-600'
                          : 'bg-orange-100 text-orange-600'
                  ]"
                >
                  <Icon
                    v-if="isBalanceType(item.type)"
                    name="dollar"
                    size="md"
                  />
                  <Icon
                    v-else-if="isSubscriptionType(item.type)"
                    name="badge"
                    size="md"
                  />
                  <Icon
                    v-else
                    name="bolt"
                    size="md"
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p class="text-xs text-gray-500">
                    {{ formatDateTime(item.used_at) }}
                  </p>
                </div>
              </div>

              <div class="text-right">
                <p
                  :class="[
                    'text-sm font-semibold',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'text-primary-600'
                        : 'text-red-600'
                      : isSubscriptionType(item.type)
                        ? 'text-purple-600'
                        : item.value >= 0
                          ? 'text-teal-600'
                          : 'text-orange-600'
                  ]"
                >
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="!isAdminAdjustment(item.type)"
                  class="font-mono text-xs text-gray-400"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
                <p v-else class="text-xs text-gray-400">
                  {{ t('redeem.adminAdjustment') }}
                </p>
                <p
                  v-if="item.notes"
                  class="mt-1 text-xs text-gray-500 italic max-w-[200px] truncate"
                  :title="item.notes"
                >
                  {{ item.notes }}
                </p>
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-else class="empty-state py-8 text-center">
            <div class="mb-4 mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 text-gray-400">
              <Icon name="clock" size="xl" />
            </div>
            <p class="text-sm text-gray-500">
              {{ t('redeem.historyWillAppear') }}
            </p>
          </div>

          <!-- 分页器 -->
          <div class="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-gray-600">
            <span>{{ t('common.total') }}: {{ historyTotal }} {{ t('pagination.results') }}</span>
            <div class="flex items-center gap-2">
              <label class="flex items-center gap-1.5">
                {{ t('pagination.perPage') }}
                <select
                  v-model="historyPageSize"
                  class="rounded-lg border border-gray-200 bg-white py-1 px-2 text-xs text-gray-800"
                  :disabled="loadingHistory || submitting"
                  @change="fetchHistory(1)"
                >
                  <option v-for="size in [20, 50, 100]" :key="size" :value="size">{{ size }}</option>
                </select>
              </label>
              <button
                class="rounded-lg border border-gray-200 bg-white px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40"
                :disabled="loadingHistory || submitting || historyPage <= 1"
                @click="fetchHistory(historyPage - 1)"
              >{{ t('pagination.previous') }}</button>
              <button
                class="rounded-lg border border-gray-200 bg-white px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-40"
                :disabled="loadingHistory || submitting || historyPage * historyPageSize >= historyTotal"
                @click="fetchHistory(historyPage + 1)"
              >{{ t('pagination.next') }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, authAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)
const yunmaoShopUrl = 'https://catfk.com/shop/ISZDQBWA'

// 5 款纯净充值卡（10, 50, 100, 200, 500），一一对应云猫专属商品直达购买链接
const cards = [
  { amount: 10, price: 10, url: 'https://catfk.com/item/b10tja' },
  { amount: 50, price: 50, url: 'https://catfk.com/item/im2q5u' },
  { amount: 100, price: 100, url: 'https://catfk.com/item/h6pve7' },
  { amount: 200, price: 200, url: 'https://catfk.com/item/k9rk9g' },
  { amount: 500, price: 500, url: 'https://catfk.com/item/u0siyt' },
]

function openShop(card: (typeof cards)[number]) {
  window.open(card.url, '_blank')
  appStore.showSuccess(`已在新窗口打开云猫店铺购买 $${card.amount} 额度！付款后请将卡密粘贴在下方兑换。`)
  const redeemSection = document.getElementById('redeem-section')
  if (redeemSection) {
    redeemSection.scrollIntoView({ behavior: 'smooth' })
  }
}


const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const historyPage = ref(1)
const historyPageSize = ref(20)
const historyTotal = ref(0)
let historyRequest = 0
let loadedHistoryPageSize = 20
const contactInfo = ref('')

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async (page = 1) => {
  const request = ++historyRequest
  const pageSize = historyPageSize.value
  loadingHistory.value = true
  try {
    const result = await redeemAPI.getHistory(page, pageSize)
    if (request !== historyRequest) return
    history.value = result.items
    historyTotal.value = result.total
    historyPage.value = page
    historyPageSize.value = pageSize
    loadedHistoryPageSize = pageSize
  } catch (error) {
    if (request !== historyRequest) return
    historyPageSize.value = loadedHistoryPageSize
    appStore.showError(t('redeem.historyLoadFailed'))
    console.error('Failed to fetch history:', error)
  } finally {
    if (request === historyRequest) loadingHistory.value = false
  }
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    try {
      await authStore.refreshUser()
    } catch (error) {
      console.error('Failed to refresh user after redeem:', error)
      appStore.showWarning(t('redeem.userRefreshFailed'))
    }

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  try {
    const settings = await authAPI.getPublicSettings()
    contactInfo.value = settings.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
