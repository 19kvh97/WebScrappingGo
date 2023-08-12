package models

type IParcell interface {
}

type ParcellListener struct {
	Listener chan IParcell
	ConfigId string
}
