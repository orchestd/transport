package recaptcha

import (
	"context"
	"fmt"
	"github.com/orchestd/dependencybundler/interfaces/configuration"
	"github.com/orchestd/dependencybundler/interfaces/credentials"
	"github.com/orchestd/dependencybundler/interfaces/transport"
	"os"
)

type GoogleReCapcha struct {
	siteKey      string
	projectId    string
	apiKey       string
	host         string
	httpClient   transport.HttpClient
	minimumScore float64
}

type Recaptcha interface {
	Check(c context.Context, siteKey, action string, score float64) (Status, error)
	CheckByAction(c context.Context, siteKey, action string) (Status, error)
}

func NewRecaptcha(conf configuration.Config, cred credentials.CredentialsGetter, httpClient transport.HttpClient) Recaptcha {
	apiKey := cred.GetCredentials().GoogleApiKey
	if apiKey == "" {
		panic("GOOGLE_API_KEY not found in Credentials")
	}

	siteKey := cred.GetCredentials().RecaptchaSiteKey
	if apiKey == "" {
		panic("GOOGLE_SITE_KEY not found in Credentials")
	}

	minimumScore, err := conf.Get("googleReCaptchaMinScore").Float64()
	if err != nil {
		panic("googleReCaptchaMinScore missing from configuration")
	}

	url, err := conf.Get("googleReCaptchaUrl").String()
	if err != nil {
		panic("googleReCaptchaUrl missing from configuration")
	}

	return GoogleReCapcha{
		projectId:    os.Getenv("PROJECT_ID"),
		host:         url,
		apiKey:       apiKey,
		minimumScore: minimumScore,
		siteKey:      siteKey,
		httpClient:   httpClient,
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
	if err != nil {
		return Status{}, err
	} else {
		return assessment.getStatus(minimumScore, action), nil
	}
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

	var response recaptchaResponse
	host := fmt.Sprintf("%s%s/assessments?key=%s", r.host, r.projectId, r.apiKey)
	err := r.httpClient.ExternalPost(c, payload, host, "", &response, nil, transport.ContentTypeJSON)
	if err.GetError() != nil {
		return response, err
	} else {
		return response, nil
	}
}
