package main

type AppModelResponseMain struct {
	OutOfSync         bool   `json:"OutOfSync"`
	LastBtcHeight     int64  `json:"LastBtcHeight"`
	CryptoFingerPrint string `json:"CryptoFingerPrint"`
	CommitmentsCount  uint64 `json:"CommitmentsCount"`
	Accelerator       string `json:"Accelerator"`
	BlockHash         string `json:"BlockHash"`
	SumExistence      uint64 `json:"SumExistence"`
	SumRemaining      uint64 `json:"SumRemaining"`
	BlockHeight       uint64 `json:"BlockHeight"`
	P2WSHCount        uint64 `json:"P2WSHCount"`
	ShutdownButton    bool   `json:"ShutdownButton"`
	DecidersHardfork  bool   `json:"DecidersHardfork"`
}
