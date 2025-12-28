package telemetry

import (
	"github.com/sturdivant20/sturdr-api/include/navigation"
	"github.com/sturdivant20/sturdr-api/include/satellite"
)

type Telemetry struct {
	Navigation navigation.Navigation `json:"navigation"`
	Satellites []satellite.Satellite `json:"satellites"`
}
