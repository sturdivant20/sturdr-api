package satellite

// Satellite data type
type Satellite struct {
	Sequence  uint64  `json:"sequence"` // Foreign key
	Week      uint16  `json:"week"`
	ToW       float32 `json:"tow"`
	PRN       uint8   `json:"prn"`
	Health    uint8   `json:"health"`
	X         float32 `json:"x"`
	Y         float32 `json:"y"`
	Z         float32 `json:"z"`
	Vx        float32 `json:"vx"`
	Vy        float32 `json:"vy"`
	Vz        float32 `json:"vz"`
	Doppler   float32 `json:"doppler"`
	PSR       float32 `json:"psr"`
	ADR       float32 `json:"adr"`
	Azimuth   float32 `json:"azimuth"`
	Elevation float32 `json:"elevation"`
	CNo       float32 `json:"cno"`
	IE        float32 `json:"ie"`
	IP        float32 `json:"ip"`
	IL        float32 `json:"il"`
	QE        float32 `json:"qe"`
	QP        float32 `json:"qp"`
	QL        float32 `json:"ql"`
}

func (sv *Satellite) Args() []any {
	return []any{
		&sv.Sequence,
		&sv.Week,
		&sv.ToW,
		&sv.PRN,
		&sv.Health,
		&sv.X,
		&sv.Y,
		&sv.Z,
		&sv.Vx,
		&sv.Vy,
		&sv.Vz,
		&sv.Doppler,
		&sv.PSR,
		&sv.ADR,
		&sv.Azimuth,
		&sv.Elevation,
		&sv.CNo,
		&sv.IE,
		&sv.IP,
		&sv.IL,
		&sv.QE,
		&sv.QP,
		&sv.QL,
	}
}
