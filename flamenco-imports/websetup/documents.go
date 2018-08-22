// Copy of https://github.com/armadillica/flamenco-manager/blob/master/websetup/documents.go
// Fetched 2018-08-21
// Would be super nice if these were public so I could just import direct from the repo

package websetup

type KeyExchangeRequest struct {
	KeyHex string `json:"key"`
}
type KeyExchangeResponse struct {
	Identifier string `json:"identifier"`
}

type LinkRequiredResponse struct {
	Required  bool   `json:"link_required"`
	ServerURL string `json:"server_url,omitempty"`
}
type LinkStartResponse struct {
	Location string `json:"location"`
}

type AuthTokenResetRequest struct {
	ManagerID  string `json:"manager_id"`
	Identifier string `json:"identifier"`
	Padding    string `json:"padding"`
	HMAC       string `json:"hmac"`
}
type AuthTokenResetResponse struct {
	Token      string `json:"token"`
	ExpireTime string `json:"expire_time"` // ignored for now, so left as string and not parsed.
}

type ErrorMessage struct {
	Message string `json:"_message"`
}
