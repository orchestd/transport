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
	recaptchaSecretKey string
	conf               configuration.Config
	httpClient         transport.HttpClient
}

type Recaptcha interface {
	Check(c context.Context, siteKey, action string, score float64) (Status, error)
	CheckByAction(c context.Context, siteKey, action string) (Status, error)
}

func NewRecaptcha(conf configuration.Config, cred credentials.CredentialsGetter, httpClient transport.HttpClient) Recaptcha {
	recaptchaSecretKey := cred.GetCredentials().RecaptchaKey
	if recaptchaSecretKey == "" {
		panic("RECAPTCHA_KEY not found in Credentials")
	}

	return GoogleReCapcha{
		conf:               conf,
		recaptchaSecretKey: recaptchaSecretKey,
		httpClient:         httpClient,
	}
}

func (r GoogleReCapcha) CheckByAction(c context.Context, siteKey, action string) (Status, error) {
	score, err := r.conf.Get("googleReCapchaScore").Float64()
	if err != nil {
		return Status{}, err
	} else if score != 0 {
		return r.Check(c, siteKey, action, score)
	} else {
		return Status{IsSuccess: true}, nil
	}

}

func (r GoogleReCapcha) Check(c context.Context, siteKey, action string, score float64) (Status, error) {
	var result Status
	siteVerifyUrl, err := r.conf.Get("googleReCapchaUrl").String()
	if err != nil {
		return result, err
	}

	data := map[string]string{
		"secret":   r.recaptchaSecretKey,
		"response": siteKey,
	}
	res, err := r.httpClient.PostForm(c, siteVerifyUrl, data, nil)
	if err != nil {
		return result, err
	}

	var body SiteVerifyResponse
	if err := json.Unmarshal(res, &body); err != nil {
		return result, fmt.Errorf("Cannot Unmarshal recaptcha body response", err)
	}

	// Check recaptcha verification success.
	if !body.Success {
		result.Error = "Unsuccessful recaptcha verify request"
		return result, nil
	}

	// Check response score.
	if body.Score < score {
		result.Error = fmt.Sprintf("Lower received score (%v) than expected", score)
		return result, nil
	}

	// Check response action.
	if action != "" {
		if body.Action != action {
			result.Error = fmt.Sprintf("Mismatched recaptcha action %s", action)
			return result, nil
		}
	}
	result.IsSuccess = true
	return result, nil
}
