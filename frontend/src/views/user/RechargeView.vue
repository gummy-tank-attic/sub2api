<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <div>
        <h2 class="text-xl font-semibold text-gray-900">{{ t('payment.unifiedRechargeTitle') }}</h2>
        <p class="mt-1 text-sm text-gray-500">{{ t('payment.unifiedRechargeDescription') }}</p>
        <div v-if="channel !== 'crypto'" class="mt-3 rounded-lg border border-primary-100 bg-primary-50/70 px-4 py-3 text-sm text-primary-800">
          <span class="font-semibold">充值规则：</span>人民币购买时 1 元 = $1 站内额度；虚拟币充值同样按购买的美元额度到账，实际 USDT / USDC 支付金额按实时汇率计算。
        </div>
      </div>
      <div class="flex gap-1 rounded-xl bg-gray-100 p-1" role="tablist" :aria-label="t('payment.unifiedRechargeTitle')">
        <button v-for="choice in choices" :key="choice.key" type="button" role="tab" :aria-selected="channel === choice.key" class="flex-1 rounded-lg px-3 py-3 text-sm font-medium transition" :class="channel === choice.key ? 'bg-white text-primary-700 shadow-sm' : 'text-gray-500 hover:text-gray-900'" @click="selectChannel(choice.key)">{{ choice.label }}</button>
      </div>
      <PaymentView v-if="channel === 'crypto'" embedded />
      <RedeemView v-else embedded />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import PaymentView from './PaymentView.vue'
import RedeemView from './RedeemView.vue'
import { useAppStore } from '@/stores/app'
import { FeatureFlags, resolveFeatureFlag } from '@/utils/featureFlags'
import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
type Channel = 'cny' | 'crypto'
const paymentEnabled = computed(() => resolveFeatureFlag(appStore.cachedPublicSettings, FeatureFlags.payment))
function initialChannel(): Channel {
  if (route.query.channel === 'cny' || route.query.channel === 'code') return 'cny'
  if (route.query.channel === 'crypto' || route.query.tab === 'subscription' || route.query.resume_token || route.query.wechat_resume_token || route.query.openid || localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)) return 'crypto'
  return 'cny'
}
const channel = ref<Channel>(initialChannel())
function selectChannel(value: Channel) {
  channel.value = value
  void router.replace({ query: { ...route.query, channel: value } })
}
const choices = computed(() => [
  { key: 'cny' as Channel, label: t('payment.cnyPurchase') },
  ...(paymentEnabled.value ? [{ key: 'crypto' as Channel, label: t('payment.cryptoPurchase') }] : []),
])
watch(() => route.query, (query) => {
  if (query.channel || query.tab === 'subscription' || query.resume_token || query.wechat_resume_token || query.openid) {
    channel.value = paymentEnabled.value ? initialChannel() : 'cny'
  }
})
watch(paymentEnabled, (enabled) => { if (!enabled && channel.value === 'crypto') channel.value = 'cny' }, { immediate: true })
</script>
