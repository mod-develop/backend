package rest

type tRequestRegistration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type tRequestAuthorization struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type actionConfirmation string

const (
	actionConfirmationAccpet actionConfirmation = "accept"
	actionConfirmationReject actionConfirmation = "reject"
)

type tRequestApiManageQuestConfirmation struct {
	ID     uint               `json:"id"`
	Action actionConfirmation `json:"action"`
}

type tRequestAPISettingsAddMaster struct {
	Code string `json:"code"`
}
