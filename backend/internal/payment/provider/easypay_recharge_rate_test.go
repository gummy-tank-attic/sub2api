package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestEasyPayBepusdtRechargeRate(t *testing.T) {
	for _, mode := range []string{"qrcode", "popup"} {
		for _, method := range []string{"usdt_trc20", "usdt_bep20", "usdc_base"} {
			for _, purpose := range []string{payment.OrderTypeBalance, payment.OrderTypeSubscription, ""} {
				for _, enabled := range []string{"true", "false", ""} {
					t.Run(mode+"/"+method+"/"+purpose+"/"+enabled, func(t *testing.T) {
						var sent url.Values
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if err := r.ParseForm(); err != nil {
								t.Error(err)
							}
							sent = r.PostForm
							w.Header().Set("Content-Type", "application/json")
							_, _ = w.Write([]byte(`{"code":1,"trade_no":"test-trade","payurl":"/checkout/test"}`))
						}))
						defer server.Close()
						p := newTestEasyPay(t, server.URL)
						p.config["paymentMode"] = mode
						p.config[EasyPayBepusdtCNYRecharge] = enabled
						p.config["customMethods"] = `[{"type":"usdt_trc20","upstreamType":"usdt.trc20"},{"type":"usdt_bep20","upstreamType":"usdt.bep20"},{"type":"usdc_base","upstreamType":"usdc.base"}]`
						resp, err := p.CreatePayment(context.Background(), payment.CreatePaymentRequest{
							OrderID: "test-order", OrderType: purpose, Amount: "67.00", PaymentType: method, Subject: "Recharge",
						})
						if mode != "popup" && enabled == "true" && purpose == payment.OrderTypeBalance {
							if err == nil || sent != nil {
								t.Fatal("unsupported mode must fail before contacting the gateway")
							}
							return
						}
						if err != nil {
							t.Fatal(err)
						}
						if mode == "popup" {
							u, err := url.Parse(resp.PayURL)
							if err != nil {
								t.Fatal(err)
							}
							sent = u.Query()
						}
						wantRate, wantFiat := "", ""
						if enabled == "true" && purpose == payment.OrderTypeBalance {
							wantRate, wantFiat = "~1", "CNY"
						}
						if sent.Get("rate") != wantRate || sent.Get("fiat") != wantFiat || sent.Get("money") != "67.00" {
							t.Fatalf("incorrect quote parameters: %v", sent)
						}
						params := map[string]string{}
						for k := range sent {
							params[k] = sent.Get(k)
						}
						if !easyPayVerifySign(params, p.config["pkey"], sent.Get("sign")) {
							t.Fatal("request signature mismatch")
						}
						if wantRate != "" {
							params["rate"] = "1"
							if easyPayVerifySign(params, p.config["pkey"], sent.Get("sign")) {
								t.Fatal("rate is not covered by signature")
							}
						}
					})
				}
			}
		}
	}
}

func TestEasyPayBepusdtRechargeRejectsOtherMethods(t *testing.T) {
	p := newTestEasyPay(t, "https://invalid.example")
	p.config[EasyPayBepusdtCNYRecharge] = "true"
	p.config["paymentMode"] = "popup"
	_, err := p.CreatePayment(context.Background(), payment.CreatePaymentRequest{OrderType: payment.OrderTypeBalance, PaymentType: "alipay", Amount: "67.00"})
	if err == nil {
		t.Fatal("expected unsupported method error")
	}
}
