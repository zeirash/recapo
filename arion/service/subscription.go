package service

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	SubscriptionService interface {
		GetActivePlans() ([]response.PlanData, error)
		GetSubscriptionByShopID(shopID int) (*response.SubscriptionData, error)
		CreateTrialSubscription(shopID int) error
		Checkout(shopID, planID int) (*response.CheckoutData, error)
		HandleMidtransWebhook(payload MidtransWebhookPayload) error
		IsSubscriptionActive(shopID int) (bool, error)
	}

	ssubscription struct{}

	MidtransWebhookPayload struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
		TransactionID     string `json:"transaction_id"`
		TransactionTime   string `json:"transaction_time"`
	}

	midtransSnapRequest struct {
		TransactionDetails midtransTransactionDetails `json:"transaction_details"`
	}

	midtransTransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount int    `json:"gross_amount"`
	}

	midtransSnapResponse struct {
		Token       string `json:"token"`
		RedirectURL string `json:"redirect_url"`
	}
)

func NewSubscriptionService() SubscriptionService {
	if subscriptionStore == nil {
		subscriptionStore = store.NewSubscriptionStore()
	}
	return &ssubscription{}
}

func (s *ssubscription) GetActivePlans() ([]response.PlanData, error) {
	plans, err := subscriptionStore.GetActivePlans()
	if err != nil {
		return nil, err
	}

	result := make([]response.PlanData, 0, len(plans))
	for _, p := range plans {
		result = append(result, response.PlanData{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Description: p.Description,
			PriceIDR:    p.PriceIDR,
			MaxUsers:    p.MaxUsers,
		})
	}
	return result, nil
}

func (s *ssubscription) GetSubscriptionByShopID(shopID int) (*response.SubscriptionData, error) {
	sub, err := subscriptionStore.GetSubscriptionByShopID(shopID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, errors.New("subscription not found")
	}

	plan, err := subscriptionStore.GetPlanByID(sub.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	data := &response.SubscriptionData{
		ID:     sub.ID,
		Status: sub.Status,
		Plan: response.PlanData{
			ID:          plan.ID,
			Name:        plan.Name,
			DisplayName: plan.DisplayName,
			Description: plan.Description,
			PriceIDR:    plan.PriceIDR,
			MaxUsers:    plan.MaxUsers,
		},
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
	}

	if sub.TrialEndsAt.Valid {
		data.TrialEndsAt = &sub.TrialEndsAt.Time
	}

	return data, nil
}

func (s *ssubscription) CreateTrialSubscription(shopID int) error {
	plans, err := subscriptionStore.GetActivePlans()
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return errors.New("no active plans found")
	}

	starterPlan := plans[0]
	trialEndsAt := time.Now().AddDate(0, 0, 14)

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = subscriptionStore.CreateTrialSubscription(tx, shopID, starterPlan.ID, trialEndsAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ssubscription) Checkout(shopID, planID int) (*response.CheckoutData, error) {
	plan, err := subscriptionStore.GetPlanByID(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	sub, err := subscriptionStore.GetSubscriptionByShopID(shopID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, errors.New("subscription not found")
	}

	orderID := fmt.Sprintf("recapo-%d-%d", shopID, time.Now().Unix())

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	payment, err := subscriptionStore.CreatePayment(tx, shopID, sub.ID, planID, orderID, plan.PriceIDR)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	snapResp, err := callMidtransSnap(orderID, plan.PriceIDR)
	if err != nil {
		return nil, fmt.Errorf("midtrans snap error: %w", err)
	}

	if err := subscriptionStore.UpdatePaymentSnapInfo(payment.ID, snapResp.Token, snapResp.RedirectURL); err != nil {
		logger.WithError(err).Error("failed to update payment snap info")
	}

	return &response.CheckoutData{
		OrderID:     orderID,
		RedirectURL: snapResp.RedirectURL,
		SnapToken:   snapResp.Token,
	}, nil
}

func (s *ssubscription) HandleMidtransWebhook(payload MidtransWebhookPayload) error {
	// Verify SHA512 signature: sha512(order_id + status_code + gross_amount + server_key)
	serverKey := cfg.MidtransServerKey
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + serverKey
	h := sha512.Sum512([]byte(raw))
	expected := fmt.Sprintf("%x", h)
	if expected != payload.SignatureKey {
		return errors.New("invalid signature")
	}

	payment, err := subscriptionStore.GetPaymentByMidtransOrderID(payload.OrderID)
	if err != nil {
		return err
	}
	if payment == nil {
		return errors.New("payment not found")
	}

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch payload.TransactionStatus {
	case constant.PaymentStatusSettlement,
		constant.PaymentStatusCapture:
		if payload.TransactionStatus == constant.PaymentStatusCapture && payload.FraudStatus != "accept" {
			break
		}
		paidAt := time.Now()
		if err := subscriptionStore.UpdatePaymentSettled(tx, payment.ID, payload.TransactionID, paidAt); err != nil {
			return err
		}
		periodEnd := time.Now().AddDate(0, 1, 0)
		if err := subscriptionStore.UpdateSubscriptionStatus(tx, payment.SubscriptionID, constant.SubStatusActive, &periodEnd); err != nil {
			return err
		}
	case constant.PaymentStatusDeny,
		constant.PaymentStatusCancel,
		constant.PaymentStatusExpire,
		constant.PaymentStatusFailure:
		if err := subscriptionStore.UpdatePaymentFailed(tx, payment.ID, payload.TransactionStatus); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *ssubscription) IsSubscriptionActive(shopID int) (bool, error) {
	sub, err := subscriptionStore.GetSubscriptionByShopID(shopID)
	if err != nil {
		return false, err
	}
	if sub == nil {
		return false, nil
	}

	now := time.Now()

	switch sub.Status {
	case constant.SubStatusActive:
		return now.Before(sub.CurrentPeriodEnd), nil
	case constant.SubStatusTrialing:
		if sub.TrialEndsAt.Valid {
			return now.Before(sub.TrialEndsAt.Time), nil
		}
		return false, nil
	default:
		return false, nil
	}
}

func callMidtransSnap(orderID string, grossAmount int) (*midtransSnapResponse, error) {
	cfg := cfg
	baseURL := "https://app.sandbox.midtrans.com"
	if !cfg.MidtransSandbox {
		baseURL = "https://app.midtrans.com"
	}

	reqBody := midtransSnapRequest{
		TransactionDetails: midtransTransactionDetails{
			OrderID:     orderID,
			GrossAmount: grossAmount,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/snap/v1/transactions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(cfg.MidtransServerKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midtrans returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var snapResp midtransSnapResponse
	if err := json.Unmarshal(respBytes, &snapResp); err != nil {
		return nil, err
	}

	return &snapResp, nil
}
