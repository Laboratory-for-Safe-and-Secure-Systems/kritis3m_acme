package handlers

import (
	"fmt"
	"net/http"

	"github.com/Laboratory-for-Safe-and-Secure-Systems/kritis3m_acme/internal/api/types"
)

func newBadRequestError(detail string) *types.Problem {
	return &types.Problem{
		Type:   "urn:ietf:params:acme:error:badRequest",
		Detail: detail,
		Status: http.StatusBadRequest,
	}
}

func newInternalServerError(detail string) *types.Problem {
	return &types.Problem{
		Type:   "urn:ietf:params:acme:error:serverInternal",
		Detail: detail,
		Status: http.StatusInternalServerError,
	}
}

func newMalformedError(detail string) *types.Problem {
	return &types.Problem{
		Type:   "urn:ietf:params:acme:error:malformed",
		Detail: detail,
		Status: http.StatusBadRequest,
	}
}

func newNotFoundError(detail string, problemType string) *types.Problem {
	problemType = fmt.Sprintf("urn:ietf:params:acme:error:%s", problemType)
	return &types.Problem{
		Type:   problemType,
		Detail: detail,
		Status: http.StatusNotFound,
	}
}
