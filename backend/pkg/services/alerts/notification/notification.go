// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package notification

// Copied from https://github.com/timonwong/prometheus-webhook-dingtalk/blob/main/notifier/notification.go

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type DingTalkNotification struct {
	MessageType string                        `json:"msgtype"`
	Text        *DingTalkNotificationText     `json:"text,omitempty"`
	Link        *DingTalkNotificationLink     `json:"link,omitempty"`
	Markdown    *DingTalkNotificationMarkdown `json:"markdown,omitempty"`
}

type DingTalkNotificationText struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type DingTalkNotificationLink struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PictureURL string `json:"picUrl"`
}

type DingTalkNotificationMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingNotificationBuilder struct {
	tmpl     *Template
	titleTpl string
	textTpl  string
}

var builder *DingNotificationBuilder

func NewNotificationBuilder() (*DingNotificationBuilder, error) {
	if builder != nil {
		return builder, nil
	}
	tmpl, err := FromDefault()
	if err != nil {
		return nil, err
	}
	builder = &DingNotificationBuilder{}
	builder.tmpl = tmpl
	builder.titleTpl = `{{ template "ding.link.title" . }}`
	builder.textTpl = `{{ template "ding.link.content" . }}`
	return builder, nil
}

func (b *DingNotificationBuilder) renderTitle(data interface{}) (string, error) {
	return b.tmpl.ExecuteTextString(b.titleTpl, data)
}

func (b *DingNotificationBuilder) renderText(data interface{}) (string, error) {
	return b.tmpl.ExecuteTextString(b.textTpl, data)
}

func (b *DingNotificationBuilder) Build(m *request.ForwardToDingTalkRequest) (*DingTalkNotification, error) {
	title, err := b.renderTitle(m)
	if err != nil {
		return nil, err
	}
	content, err := b.renderText(m)
	if err != nil {
		return nil, err
	}

	notification := &DingTalkNotification{
		MessageType: "markdown",
		Markdown: &DingTalkNotificationMarkdown{
			Title: title,
			Text:  content,
		},
	}

	return notification, nil
}

func SendNotification(notification *DingTalkNotification, configUrl, secret string) error {
	targetURL, _ := url.Parse(configUrl)
	if secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
		stringToSign := []byte(timestamp + "\n" + string(secret))

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(stringToSign)
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		qs := targetURL.Query()
		qs.Set("timestamp", timestamp)
		qs.Set("sign", signature)
		targetURL.RawQuery = qs.Encode()
	}

	body, err := json.Marshal(&notification)
	if err != nil {
		return err
	}
	client := http.Client{}
	httpReq, err := http.NewRequest("POST", targetURL.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return errors.Errorf("unacceptable response code %d", resp.StatusCode)
	}
	return nil
}
