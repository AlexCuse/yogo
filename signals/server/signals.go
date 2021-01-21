package server

type Signal struct {
	Name   string `gorm:"primaryKey;autoIncrement:false" json:"name,omitempty"`
	Source string `json:"source,omitempty"`
}

type SignalWithHitCount struct {
	Signal
	Count int `json:"count,omitempty"`
}
