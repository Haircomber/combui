package main

type Wall struct {
	loader JQuery

	name JQuery
	addr JQuery
	bal  JQuery
	rej  JQuery
	img  JQuery

	syncbtn          JQuery
	reloadbtn        JQuery
	loadbtn          JQuery
	stealthtable     JQuery
	stealthpaginator JQuery

	tx []*AppModelResponseBid
}

func NewWall() *Wall {
	return &Wall{jQuery("#loader"),
		jQuery("#paywall-recipient"),
		jQuery("#paywall-address"),
		jQuery("#paywall-balance"),
		jQuery("#paywall-rejected"),
		jQuery("#paywall-wallet-image"),
		jQuery("#paywall-sync-btn"),
		jQuery("#paywall-reload-btn"),
		jQuery("#paywall-load-btn"),
		jQuery("#paywall-stealths-keys"),
		jQuery("#paywall-stealths-paginator"),
		nil,
	}
}
