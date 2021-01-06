package contracts

type Scannable struct {
	Movement

	// precalculated fields to support DSL more easily
	AbsoluteChange float64
	PercentChange  float64
	AverageVolume  int
}
