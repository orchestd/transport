package recaptcha

import (
	"cloud.google.com/go/recaptchaenterprise/v2/apiv1beta1/recaptchaenterprisepb"
	"context"
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
	recaptchaSecretKey string
	siteKey            string
	projectID          string
	apiKey             string
	host               string
	conf               configuration.Config
	httpClient         transport.HttpClient
	minimumScore       float64
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
	minimumScore, err := conf.Get("googleReCaptchaMinScore").Float64()
	if err != nil {
		panic("googleReCaptchaMinScore missing from configuration")
	}

	googleReCapchaUrl, err := conf.Get("googleReCaptchaUrl").String()
	if err != nil {
		panic("googleReCaptchaUrl missing from configuration")
	}

	return GoogleReCapcha{
		conf:               conf,
		recaptchaSecretKey: recaptchaSecretKey,
		minimumScore:       minimumScore,
		siteKey:            googleReCapchaUrl,
		httpClient:         httpClient,
	}
}

func (r GoogleReCapcha) CheckByAction(c context.Context, token, action string) (Status, error) {
	if r.minimumScore != 0 {
		return r.Check(c, token, action, r.minimumScore)
	} else {
		return Status{IsSuccess: true}, nil
	}
}

func (r GoogleReCapcha) Check(c context.Context, token, action string, minimumScore float64) (Status, error) {
	assessment, err := r.createAssessment(c, token, action)
	return assessment.getStatus(float32(minimumScore), action), err
}

func (r GoogleReCapcha) createAssessment(c context.Context, token string, recaptchaAction string) (recaptchaResponse, error) {
	event := map[string]string{
		"token":          token,
		"expectedAction": recaptchaAction,
		"siteKey":        r.siteKey,
	}
	payload := map[string]map[string]string{
		"event": event,
	}

	var response recaptchaenterprisepb.Assessment
	host := fmt.Sprintf("%s%s/assesments?key=%s", r.host, r.projectID, r.apiKey)
	// "https://recaptchaenterprise.googleapis.com/v1/projects/"+r.projectID+"/assessments?key="+"AIz.....61w"
	err := r.httpClient.ExternalPost(c, payload, host, "", &response, nil, transport.ContentTypeJSON)
	return recaptchaResponse(response), err
}
