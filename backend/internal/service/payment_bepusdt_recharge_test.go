package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/payment/provider"
)

func TestBepusdtRechargeMultiplierGuard(t *testing.T) {
	for _, tc := range []struct {
		name, purpose, providerKey, enabled string
		multiplier                          float64
		wantError                           bool
	}{
		{"ready", payment.OrderTypeBalance, payment.TypeEasyPay, "true", 1, false},
		{"old-fixed-rate", payment.OrderTypeBalance, payment.TypeEasyPay, "true", 7, true},
		{"double-conversion", payment.OrderTypeBalance, payment.TypeEasyPay, "true", 6.7, true},
		{"subscription", payment.OrderTypeSubscription, payment.TypeEasyPay, "true", 7, false},
		{"not-enabled", payment.OrderTypeBalance, payment.TypeEasyPay, "", 7, false},
		{"other-provider", payment.OrderTypeBalance, payment.TypeStripe, "true", 7, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sel := &payment.InstanceSelection{ProviderKey: tc.providerKey, Config: map[string]string{provider.EasyPayBepusdtCNYRecharge: tc.enabled}}
			err := validateBepusdtRechargeMultiplier(tc.purpose, &PaymentConfig{BalanceRechargeMultiplier: tc.multiplier}, sel)
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tc.wantError)
			}
		})
	}
}

func TestBepusdtRechargeOrderPurposeForwarded(t *testing.T) {
	for _, purpose := range []string{payment.OrderTypeBalance, payment.OrderTypeSubscription} {
		got := buildProviderCreatePaymentRequest(CreateOrderRequest{OrderType: purpose}, nil, "order", "67.00", "subject")
		if got.OrderType != purpose || got.Amount != "67.00" {
			t.Fatalf("unexpected provider request: %+v", got)
		}
	}
}

func TestBepusdtRechargeSwitchProtectedForPendingOrders(t *testing.T) {
	oldConfig := map[string]string{}
	newConfig := map[string]string{provider.EasyPayBepusdtCNYRecharge: "true"}
	if !hasPendingOrderProtectedConfigChange(payment.TypeEasyPay, oldConfig, newConfig) ||
		!hasPendingOrderProtectedConfigChange(payment.TypeEasyPay, newConfig, oldConfig) {
		t.Fatal("changing the quote policy must use the pending-order guard")
	}
}
