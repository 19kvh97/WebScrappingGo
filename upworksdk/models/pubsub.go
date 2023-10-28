package models

type PackageType int

const (
	UPWORK_JOB_PACKAGE PackageType = iota
)

type Package struct {
	Type PackageType
	Data IParcell
}

type IDistributor interface {
	Update()
}

type Distributor struct {
	IDistributor
	ID      int
	Channel chan IParcell
}
