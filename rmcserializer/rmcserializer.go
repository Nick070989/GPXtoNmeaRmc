package rmcserializer

import (
	"fmt"
	"time"
)

type RmcSerializer struct {
	latitude  float64
	longitude float64
	time      time.Time
	validity  bool
	course    float64
	speed     float64
}

func New(lat, long, course, speeed float64,
	validity bool, time time.Time) *RmcSerializer {
	return &RmcSerializer{latitude: lat, longitude: long,
		time: time, validity: validity, course: course, speed: speeed}
}

func calculateChecksum(sentence string) string {
	var checksum byte = 0
	for i := 0; i < len(sentence); i++ {
		checksum ^= sentence[i]
	}
	return fmt.Sprintf("%02X", checksum)
}

func formatNmeaLatitude(lat float64) string {
	deg := int(lat)
	min := (lat - float64(deg)) * 60
	return fmt.Sprintf("%02d%08.4f", deg, min)
}

func formatNmeaLongitude(lon float64) string {
	deg := int(lon)
	min := (lon - float64(deg)) * 60
	return fmt.Sprintf("%03d%08.4f", deg, min)
}

func (s *RmcSerializer) Serialize() string {
	timeStr := s.time.Format("150405.00")

	dateStr := s.time.Format("020106")

	validityStr := "A"
	if !s.validity {
		validityStr = "V"
	}

	latDir := "N"
	if s.latitude < 0 {
		latDir = "S"
		s.latitude = -s.latitude
	}
	latStr := formatNmeaLatitude(s.latitude)

	lonDir := "E"
	if s.longitude < 0 {
		lonDir = "W"
		s.longitude = -s.longitude
	}
	lonStr := formatNmeaLongitude(s.longitude)

	speed := fmt.Sprintf("%.1f", s.speed)

	course := fmt.Sprintf("%.1f", s.course)

	baseBody := fmt.Sprintf("GPRMC,%s,%s,%s,%s,%s,%s,%s,%s,%s,,,A",
		timeStr, validityStr, latStr, latDir, lonStr, lonDir, speed, course, dateStr)

	checksum := calculateChecksum(baseBody)

	return fmt.Sprintf("$%s*%s\r\n", baseBody, checksum)
}
