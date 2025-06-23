package main

type AppModelResponseBids struct {
	Tx []*AppModelResponseBid `json:"Tx"`
}
type AppModelResponseBid struct {
	TxId        string                       `json:"TxId"`
	TxOut       []*AppModelResponseBidCommit `json:"TxOut"`
	WitnessTxId string                       `json:"WitnessTxId"`
	Weight      uint64                       `json:"Weight"`
	Size        uint64                       `json:"Size"`
	Fee         uint64                       `json:"Fee"`
	FeeSizeKb   uint64                       `json:"FeeSizeKb"`
	FeeWeightKb uint64                       `json:"FeeWeightKb"`
	Testnet     bool                         `json:"Testnet"`
}

type AppModelResponseBidCommit struct {
	Commitment string `json:"Commitment"`
}
