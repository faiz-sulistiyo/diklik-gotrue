package sms_provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/supabase/gotrue/internal/conf"
	"github.com/supabase/gotrue/internal/utilities"
)

const (
	defaultYMApiBase = "https://r2.cloud.yellow.ai/api/engagements/notifications/v2/push?bot=x1688370326056"
)

type YMProvider struct {
	Config  *conf.YMProviderConfiguration
	APIPath string
}

type YMResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	MsgID   bool   `json:"msg_id"`
}

type YMErrResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

type YMRequest struct {
	UserDetails  YMUserDetails  `json:"userDetails"`
	Notification YMNotification `json:"notification"`
	Config       YMConfig       `json:"config"`
}

type YMUserDetails struct {
	Number string `json:"number"`
}

type YMNotification struct {
	Type       string   `json:"type"`
	Sender     string   `json:"sender"`
	Language   string   `json:"language"`
	Namespace  string   `json:"namespace"`
	TemplateID string   `json:"templateId"`
	Params     YMParams `json:"params"`
}

type YMConfig struct {
	ScheduleAt string `json:"scheduleAt"`
}

type YMParams struct {
	First string `json:"1"`
}

// Creates a SmsProvider with the YM Config
func NewYmProvider(config conf.YMProviderConfiguration) (SmsProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	apiPath := defaultYMApiBase
	return &YMProvider{
		Config:  &config,
		APIPath: apiPath,
	}, nil
}

func (t *YMProvider) SendMessage(phone string, message string, channel string) (string, error) {
	if channel == WhatsappProvider {
		return "", t.SendWhatsapp(phone, message, channel)
	}
	return "", fmt.Errorf("channel type %q is not supported for YM", channel)
}

// Send an WA containing the OTP with YM's API
func (t *YMProvider) SendWhatsapp(phone, message, channel string) error {
	sender := t.Config.MessageServiceSid

	body := &YMRequest{
		UserDetails:  YMUserDetails{Number: phone},
		Notification: YMNotification{Type: "whatsapp", Sender: sender, Language: "id", Namespace: t.Config.Namespace, TemplateID: "kirim_pin", Params: YMParams{First: message}},
		Config:       YMConfig{ScheduleAt: "immediate"},
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(body)

	client := &http.Client{Timeout: defaultTimeout}

	r, err := http.NewRequest("POST", t.APIPath, payloadBuf)

	if err != nil {
		return err
	}

	r.Header.Add("x-api-key", t.Config.AccountSid)
	r.Header.Add("Content-Type", "application/json")

	res, err := client.Do(r)

	defer utilities.SafeClose(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		resp := &YMErrResponse{}
		if err := json.NewDecoder(res.Body).Decode(resp); err != nil {
			return err
		}
		return err
	}

	return nil
}
