package realtime

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	AppID  string
	Key    string
	Secret string
	Host   string
	Port   int
	Secure bool
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.Host == "" {
		cfg.Host = "courseexe.com"
	}
	if cfg.Port == 0 {
		cfg.Port = 443
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type TriggerPayload struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
	Data     string   `json:"data"`
	SocketID string   `json:"socket_id,omitempty"`
}

func (c *Client) Trigger(channel, event string, data interface{}) error {
	return c.TriggerMulti([]string{channel}, event, data, "")
}

func (c *Client) TriggerMulti(channels []string, event string, data interface{}, socketID string) error {
	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		dataStr = string(b)
	}

	payload := TriggerPayload{
		Name:     event,
		Channels: channels,
		Data:     dataStr,
		SocketID: socketID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Calculate MD5 of request body
	md5Hash := md5.Sum(bodyBytes)
	bodyMD5 := hex.EncodeToString(md5Hash[:])

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Canonical signature string
	canonical := fmt.Sprintf("POST\n/apps/%s/events\nauth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		c.cfg.AppID, c.cfg.Key, timestamp, bodyMD5)

	h := hmac.New(sha256.New, []byte(c.cfg.Secret))
	h.Write([]byte(canonical))
	authSignature := hex.EncodeToString(h.Sum(nil))

	proto := "https"
	if !c.cfg.Secure {
		proto = "http"
	}
	portSuffix := ""
	if (c.cfg.Secure && c.cfg.Port != 443) || (!c.cfg.Secure && c.cfg.Port != 80) {
		portSuffix = fmt.Sprintf(":%d", c.cfg.Port)
	}

	q := url.Values{}
	q.Set("auth_key", c.cfg.Key)
	q.Set("auth_timestamp", timestamp)
	q.Set("auth_version", "1.0")
	q.Set("body_md5", bodyMD5)
	q.Set("auth_signature", authSignature)

	endpoint := fmt.Sprintf("%s://%s%s/apps/%s/events?%s", proto, c.cfg.Host, portSuffix, c.cfg.AppID, q.Encode())

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) AuthenticateChannel(socketID, channelName string, channelData *string) map[string]string {
	toSign := fmt.Sprintf("%s:%s", socketID, channelName)
	if channelData != nil {
		toSign = fmt.Sprintf("%s:%s", toSign, *channelData)
	}

	h := hmac.New(sha256.New, []byte(c.cfg.Secret))
	h.Write([]byte(toSign))
	sig := hex.EncodeToString(h.Sum(nil))

	result := map[string]string{
		"auth": fmt.Sprintf("%s:%s", c.cfg.Key, sig),
	}
	if channelData != nil {
		result["channel_data"] = *channelData
	}
	return result
}
