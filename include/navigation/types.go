package navigation

// Navigation datatype
type Navigation struct {
	Sequence  uint64  `json:"sequence"`
	Week      uint16  `json:"week"`
	ToW       float32 `json:"tow"`
	NSat      uint8   `json:"n_sat"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Altitude  float32 `json:"altitude"`
	Vn        float32 `json:"vn"`
	Ve        float32 `json:"ve"`
	Vd        float32 `json:"vd"`
	Roll      float32 `json:"roll"`
	Pitch     float32 `json:"pitch"`
	Yaw       float32 `json:"yaw"`
	PDOP      float32 `json:"pdop"`
	HDOP      float32 `json:"hdop"`
	VDOP      float32 `json:"vdop"`
}

func (n *Navigation) Args() []any {
	return []any{
		&n.Sequence,
		&n.Week,
		&n.ToW,
		&n.NSat,
		&n.Latitude,
		&n.Longitude,
		&n.Altitude,
		&n.Vn,
		&n.Ve,
		&n.Vd,
		&n.Roll,
		&n.Pitch,
		&n.Yaw,
		&n.PDOP,
		&n.HDOP,
		&n.VDOP,
	}
}
