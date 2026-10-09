import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount, enableAutoUnmount } from '@vue/test-utils'
import { reactive } from 'vue'
import RechargeView from '../RechargeView.vue'
import PaymentView from '../PaymentView.vue'
import RedeemView from '../RedeemView.vue'
import { PAYMENT_RECOVERY_STORAGE_KEY } from '@/components/payment/paymentFlow'

const state = vi.hoisted(() => ({ settings: { payment_enabled: true }, route: { query: {} as Record<string, string> } }))
const replace = vi.hoisted(() => vi.fn().mockResolvedValue(undefined))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ cachedPublicSettings: state.settings }) }))
vi.mock('vue-router', () => ({ useRoute: () => state.route, useRouter: () => ({ replace }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<div><slot /></div>' } }))
vi.mock('../PaymentView.vue', () => ({ default: { name: 'PaymentView', props: ['embedded'], template: '<div />' } }))
vi.mock('../RedeemView.vue', () => ({ default: { name: 'RedeemView', props: ['embedded', 'codeOnly'], template: '<div />' } }))
enableAutoUnmount(afterEach)
beforeEach(() => {
  state.settings = reactive({ payment_enabled: true })
  state.route = reactive({ query: {} })
  localStorage.clear()
  replace.mockClear()
})
function mountPage() {
  return shallowMount(RechargeView, { global: { stubs: { AppLayout: { template: '<div><slot /></div>' } } } })
}
describe('unified recharge choices', () => {
  it('opens CNY purchase and switches to crypto', async () => {
    const wrapper = mountPage()
    expect(wrapper.getComponent(RedeemView).exists()).toBe(true)
    await wrapper.findAll('[role="tab"]')[1].trigger('click')
    expect(wrapper.findComponent(PaymentView).exists()).toBe(true)
    expect(wrapper.findAll('[role="tab"]')).toHaveLength(2)
    expect(replace).toHaveBeenCalledWith({ query: { channel: 'crypto' } })
  })
  it('keeps CNY purchase and code redemption accessible when payment is disabled', async () => {
    state.route.query = { channel: 'crypto' }
    state.settings.payment_enabled = false
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.findAll('[role="tab"]')).toHaveLength(1)
    expect(wrapper.findComponent(PaymentView).exists()).toBe(false)
    expect(wrapper.findComponent(RedeemView).exists()).toBe(true)
  })
  it('opens the payment flow directly for subscription renewal links', () => {
    state.route.query = { tab: 'subscription', group: '2' }
    expect(mountPage().findComponent(PaymentView).exists()).toBe(true)
  })
  it('opens the payment flow for a saved unpaid order', () => {
    localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, '{"orderId":123}')
    expect(mountPage().findComponent(PaymentView).exists()).toBe(true)
  })
  it('keeps an explicit CNY or code link accessible with a saved unpaid order', () => {
    localStorage.setItem(PAYMENT_RECOVERY_STORAGE_KEY, '{"orderId":123}')
    state.route.query = { channel: 'code' }
    expect(mountPage().findComponent(RedeemView).exists()).toBe(true)
  })
  it('keeps the payment view mounted when callback parameters are cleared', async () => {
    state.route.query = { wechat_resume_token: 'resume-123' }
    const wrapper = mountPage()
    const payment = wrapper.getComponent(PaymentView).vm
    state.route.query = {}
    await flushPromises()
    expect(wrapper.getComponent(PaymentView).vm).toBe(payment)
  })
})
