package recaptcha

import (
	"cloud.google.com/go/recaptchaenterprise/v2/apiv1beta1/recaptchaenterprisepb"
	"fmt"
)

type Status struct {
	IsSuccess bool
	Error     string
}

type recaptchaResponse recaptchaenterprisepb.Assessment

func (response *recaptchaResponse) getStatus(minimumScore float32, action string) Status {
	var status Status
	// Check recaptcha verification success.
	//if !response.Success {
	//	status.Error = "Unsuccessful recaptcha verify request"
	//	return status
	//}
	if !response.TokenProperties.Valid {
		status.Error = response.TokenProperties.InvalidReason.String()
		return status
	}

	// Check response score.
	if response.Score < minimumScore {
		status.Error = fmt.Sprintf("Lower received score (%v) than expected minimum score (%v)", response.Score, minimumScore)
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
