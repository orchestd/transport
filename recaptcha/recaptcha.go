package recaptcha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const siteVerifyURL = "https://www.google.com/recaptcha/api/siteverify"
const score = 0.5

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

func CheckRecaptcha(recaptchaSecretKey, response, action string) error {
	req, err := http.NewRequest(http.MethodPost, siteVerifyURL, nil)
	if err != nil {
		return fmt.Errorf("couldn't create recaptcha NewRequest to siteverify ", err)
	}

	// Add necessary request parameters.
	q := req.URL.Query()
	q.Add("secret", recaptchaSecretKey)
	q.Add("response", response)
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't make request to check recaptcha", err)
	}
	defer resp.Body.Close()

	// Decode response.
	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("couldn't decode recaptcha body response", err)
	}

	// Check recaptcha verification success.
	if !body.Success {
		return fmt.Errorf("unsuccessful recaptcha verify request", nil, "wrongUNPW")
	}

	// Check response score.
	if body.Score < score {
		return fmt.Errorf("lower received score than expected", nil, "wrongUNPW")
	}

	// Check response action.
	if body.Action != action {
		return fmt.Errorf("mismatched recaptcha action", nil, "wrongUNPW")
	}

	return nil
}
