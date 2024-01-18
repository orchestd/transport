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

type Status struct {
	IsSuccess bool
	Error     string
}

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
	Check(c context.Context, siteKey, action string, score float64) (Status, error)
	CheckByAction(c context.Context, siteKey, action string) (Status, error)
}

func NewRecaptcha(conf configuration.Config, cred credentials.CredentialsGetter, httpClient transport.HttpClient) Recaptcha {
	return GoogleReCapcha{
		conf:       conf,
		cred:       cred,
		httpClient: httpClient,
	}
}

func (r GoogleReCapcha) CheckByAction(c context.Context, siteKey, action string) (Status, error) {
	return r.Check(c, siteKey, action, 0.5)
}

func (r GoogleReCapcha) Check(c context.Context, siteKey, action string, score float64) (Status, error) {
	var result Status
	siteVerifyUrl, err := r.conf.Get("siteverifyurl").String()
	if err != nil {
		return result, err
	}

	recaptchaSecretKey := r.cred.GetCredentials().RecaptchaKey
	if recaptchaSecretKey == "" {
		return result, fmt.Errorf("RECAPTCHA_KEY not found in Credentials")
	}
	data := map[string]string{
		"secret":   recaptchaSecretKey,
		"response": siteKey,
	}
	res, err := r.httpClient.PostForm(c, siteVerifyUrl, data, nil)
	if err != nil {
		return result, err
	}

	var body SiteVerifyResponse
	if err := json.Unmarshal(res, &body); err != nil {
		return result, fmt.Errorf("couldn't decode recaptcha body response", err)
	}

	// Check recaptcha verification success.
	if !body.Success {
		result.Error = "unsuccessful recaptcha verify request"
		return result, nil
	}

	// Check response score.
	if body.Score < score {
		result.Error = "lower received score than expected"
		return result, nil
	}

	// Check response action.
	if action != "" {
		if body.Action != action {
			result.Error = "mismatched recaptcha action"
			return result, nil
		}
	}
	result.IsSuccess = true
	return result, nil
}
