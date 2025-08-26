package gardener

import (
	"fmt"
	"strconv"
	"strings"

	gardener "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerhelper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
)

type ErrReason string

func IsRetryable(lastErrors []gardener.LastError) bool {
	if len(lastErrors) == 0 {
		return false
	}

	// we can't retry those
	if hasOneOfGardenerNonRetryableErrorCode(lastErrors...) {
		return false
	}

	// we can retry those
	if hasOneOfKnownRetryableGardenerErrorCode(lastErrors...) {
		return true
	}

	// adds resiliency for any possible new retryable error codes
	if !gardenerhelper.HasNonRetryableErrorCode(lastErrors...) {
		return true
	}

	return false
}

func hasOneOfGardenerNonRetryableErrorCode(lastErrors ...gardener.LastError) bool {
	for _, lastError := range lastErrors {
		for _, code := range lastError.Codes {
			if code == gardener.ErrorInfraUnauthenticated ||
				code == gardener.ErrorInfraUnauthorized {
				return true
			}
		}
	}
	return false
}

func hasOneOfKnownRetryableGardenerErrorCode(lastErrors ...gardener.LastError) bool {
	for _, lastError := range lastErrors {
		for _, code := range lastError.Codes {
			if code == gardener.ErrorInfraRateLimitsExceeded ||
				code == gardener.ErrorInfraQuotaExceeded ||
				code == gardener.ErrorInfraDependencies ||
				code == gardener.ErrorConfigurationProblem ||
				code == gardener.ErrorProblematicWebhook {
				return true
			}
		}
	}
	return false
}

func ToErrReason(lastErrors ...gardener.LastError) ErrReason {
	var codes []gardener.ErrorCode
	var vals []string

	for _, e := range lastErrors {
		if len(e.Codes) > 0 {
			codes = append(codes, e.Codes...)
		}
	}

	for _, code := range codes {
		vals = append(vals, string(code))
	}

	return ErrReason(strings.Join(vals, ", "))
}

func CombineErrorDescriptions(lastErrors []gardener.LastError) string {
	var descriptions string
	for i, lastError := range lastErrors {
		descriptions += fmt.Sprint(strconv.Itoa(i+1), ") ", lastError.Description, " ")
	}
	return descriptions
}
