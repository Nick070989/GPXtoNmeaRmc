package gpxreader

import (
	"errors"
	"io"

	"github.com/thcyron/gpx"
)

var trackHasNoPointsError = errors.New("Track has no points!")

type GpxReader struct {
	gpxPoints []gpx.Point
}

func (g GpxReader) GetPoints(input io.Reader) ([]gpx.Point, error) {
	doc, err := gpx.NewDecoder(input).Decode()
	if err != nil {
		return nil, err
	}

	for _, trk := range doc.Tracks {
		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				// if pt.Time.IsZero() {
				// 	continue
				// }
				g.gpxPoints = append(g.gpxPoints, pt)
			}
		}
	}
	if len(g.gpxPoints) == 0 {
		return nil, trackHasNoPointsError
	}
	return g.gpxPoints, nil
}
