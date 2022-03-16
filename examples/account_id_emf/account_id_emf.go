package account_id_emf

import (
	"github.com/glassechidna/go-emf/emf"
	"github.com/glassechidna/go-emf/emf/unit"
	"net/http"
)

type AccountIdEmf struct{}

func (a *AccountIdEmf) OnRequest(req *http.Request) {
	auth := req.Header.Get("Authorization")

	parsed, ok := ParseAuthorizationHeader(auth)
	if !ok {
		return
	}

	emf.Namespace = "flowdog"
	emf.Emit(emf.MSI{
		"KeyId":     parsed.KeyId,
		"Date":      parsed.Date,
		"Region":    parsed.Region,
		"Service":   parsed.Service,
		"AccountId": emf.Dimension(parsed.AccountId),
		"Requests":  emf.Metric(1, unit.Count),
	})
}

func (a *AccountIdEmf) OnResponse(resp *http.Response) error {
	return nil
}
