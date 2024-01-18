package recaptcha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orchestd/dependencybundler/interfaces/configuration"
	"github.com/orchestd/dependencybundler/interfaces/credentials"
	"github.com/orchestd/dependencybundler/interfaces/transport"
	"time"
)

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type GoogleReCapcha struct {
	conf       configuration.Config
	cred       credentials.CredentialsGetter
	httpClient transport.HttpClient
}

type Recaptcha interface {
	Check(c context.Context, siteKey, action string, score float64) ([]string, error)
	CheckByAction(c context.Context, siteKey, action string) ([]string, error)
}

func NewRecaptcha(conf configuration.Config, cred credentials.CredentialsGetter, httpClient transport.HttpClient) Recaptcha {
	return GoogleReCapcha{
		conf:       conf,
		cred:       cred,
		httpClient: httpClient,
	}
}

func (r GoogleReCapcha) CheckByAction(c context.Context, siteKey, action string) ([]string, error) {
	return r.Check(c, siteKey, action, 0.5)
}

func (r GoogleReCapcha) Check(c context.Context, siteKey, action string, score float64) ([]string, error) {
	siteVerifyUrl, err := r.conf.Get("siteverifyurl").String()
	if err != nil {
		return []string{}, err
	}

	recaptchaSecretKey := r.cred.GetCredentials().RecaptchaKey
	if recaptchaSecretKey == "" {
		return []string{}, fmt.Errorf("RECAPTCHA_KEY not found in Credentials")
	}
	data := map[string]string{
		"secret":   recaptchaSecretKey,
		"response": siteKey,
	}
	res, err := r.httpClient.PostForm(c, siteVerifyUrl, data, nil)
	if err != nil {
		return []string{}, err
	}

	var body SiteVerifyResponse
	if err := json.Unmarshal(res, &body); err != nil {
		return []string{}, fmt.Errorf("couldn't decode recaptcha body response", err)
	}

	// Check recaptcha verification success.
	if !body.Success {
		return body.ErrorCodes, fmt.Errorf("unsuccessful recaptcha verify request", nil, "wrongUNPW")
	}

	// Check response score.
	if body.Score < score {
		return body.ErrorCodes, fmt.Errorf("lower received score than expected", nil, "wrongUNPW")
	}

	// Check response action.
	if action != "" {
		if body.Action != action {
			return body.ErrorCodes, fmt.Errorf("mismatched recaptcha action", nil, "wrongUNPW")
		}
	}
	return []string{}, nil
}
