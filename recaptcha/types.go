package recaptcha

import (
	"fmt"
	"time"
)

type Status struct {
	IsSuccess bool
	Error     string
}

type event struct {
	Token                    string        `json:"token"`
	SiteKey                  string        `json:"siteKey"`
	UserAgent                string        `json:"userAgent"`
	UserIpAddress            string        `json:"userIpAddress"`
	ExpectedAction           string        `json:"expectedAction"`
	HashedAccountId          string        `json:"hashedAccountId"`
	Express                  bool          `json:"express"`
	RequestedUri             string        `json:"requestedUri"`
	WafTokenAssessment       bool          `json:"wafTokenAssessment"`
	Ja3                      string        `json:"ja3"`
	Ja4                      string        `json:"ja4"`
	Headers                  []interface{} `json:"headers"`
	FirewallPolicyEvaluation bool          `json:"firewallPolicyEvaluation"`
	FraudPrevention          string        `json:"fraudPrevention"`
}

type riskAnalysis struct {
	Score                  float64       `json:"score"`
	Reasons                []interface{} `json:"reasons"`
	ExtendedVerdictReasons []interface{} `json:"extendedVerdictReasons"`
	Challenge              string        `json:"challenge"`
	VerifiedBots           []interface{} `json:"verifiedBots"`
}

type tokenProperties struct {
	Valid              bool      `json:"valid"`
	InvalidReason      string    `json:"invalidReason"`
	Hostname           string    `json:"hostname"`
	AndroidPackageName string    `json:"androidPackageName"`
	IosBundleId        string    `json:"iosBundleId"`
	Action             string    `json:"action"`
	CreateTime         time.Time `json:"createTime"`
}

type accountDefenderAssessment struct {
	Labels []interface{} `json:"labels"`
}

type recaptchaResponse struct {
	Name                      string                    `json:"name"`
	Event                     event                     `json:"event"`
	RiskAnalysis              riskAnalysis              `json:"riskAnalysis"`
	TokenProperties           tokenProperties           `json:"tokenProperties"`
	AccountDefenderAssessment accountDefenderAssessment `json:"accountDefenderAssessment"`
}

func (response *recaptchaResponse) getStatus(minimumScore float64, action string) Status {
	var status Status
	// Check recaptcha verification success.
	//if !response.Success {
	//	status.Error = "Unsuccessful recaptcha verify request"
	//	return status
	//}
	if !response.TokenProperties.Valid {
		status.Error = response.TokenProperties.InvalidReason
		return status
	}

	// Check response score.
	if response.RiskAnalysis.Score < minimumScore {
		status.Error = fmt.Sprintf("Lower received score (%v) than expected minimum score (%v)", response.RiskAnalysis.Score, minimumScore)
		return status
	}

	// Check response action.
	if action != "" {
		if response.Event.ExpectedAction != action {
			status.Error = fmt.Sprintf("Mismatched recaptcha action %s", action)
			return status
		}
	}

	status.IsSuccess = true
	return status
}
