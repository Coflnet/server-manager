package model

type PaymentPayload struct {
	Id           string `json:"id"`
	UserId       string `json:"userId"`
	ProductSlug  string `json:"productSlug"`
	ProductId    string `json:"productId"`
	OwnedSeconds int    `json:"ownedSeconds"`
	Amount       string `json:"amount"`
	Reference    string `json:"reference"`
	Timestamp    string `json:"timestamp"`
}
