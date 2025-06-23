package main

type AppModelResponseRange struct {
	Combbases []ResponseCombbase `json:"Combbases"`
}
type ResponseCombbase struct {
	Combbase string `json:"Combbase"`
	Height   uint64 `json:"Height"`
}