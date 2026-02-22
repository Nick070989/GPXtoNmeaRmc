package main

import (
	"flag"
	"fmt"
	"gpxtormc/gpxreader"
	"gpxtormc/rmcserializer"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/govitia/geocalc"
	"github.com/thcyron/gpx"
	"go.bug.st/serial"
)

const metrPerHourToKnots float64 = 1.94384449244061

func CalcCourse(pCurrent gpx.Point, pNext gpx.Point) float64 {
	p1 := geocalc.NewPoint(
		geocalc.DegreeToRad(pCurrent.Latitude),
		geocalc.DegreeToRad(pCurrent.Longitude),
	)
	p2 := geocalc.NewPoint(
		geocalc.DegreeToRad(pNext.Latitude),
		geocalc.DegreeToRad(pNext.Longitude),
	)
	bearingRad := p1.Bearing(p2)
	bearingDeg := geocalc.RadToDegree(bearingRad)
	if bearingDeg < 0 {
		bearingDeg += 360
	}
	return bearingDeg
}

func CalcSpeedInKnots(pCurrent gpx.Point, pPrevius gpx.Point) float64 {
	p1 := geocalc.NewPoint(
		geocalc.DegreeToRad(pCurrent.Latitude),
		geocalc.DegreeToRad(pCurrent.Longitude),
	)
	p2 := geocalc.NewPoint(
		geocalc.DegreeToRad(pPrevius.Latitude),
		geocalc.DegreeToRad(pPrevius.Longitude),
	)
	dist := geocalc.Distance(p1, p2)
	seconds := pCurrent.Time.Unix() - pPrevius.Time.Unix()
	if seconds <= 0 {
		return 0.0
	}
	return dist / float64(seconds) * metrPerHourToKnots
}

func sendPointsToPort(points []gpx.Point, port serial.Port) {
	if port == nil {
		return
	}
	var rmcList []string
	for i, point := range points {
		speed := 0.0
		if i != 0 {
			speed = CalcSpeedInKnots(point, points[i-1])
		}
		course := 0.0
		if i != len(points)-1 {
			course = CalcCourse(point, points[i+1])
		}
		rmc := rmcserializer.New(point.Latitude, point.Longitude,
			course, speed, true, point.Time)
		rmcStr := rmc.Serialize()
		if rmcStr == "" {
			println("ERROR!", "Bad serialize")
			continue
		}
		_, err := port.Write([]byte(rmcStr))
		if err != nil {
			fmt.Printf("Send message error: %v\n", err)
		}
		if i < len(points)-1 {
			seconds := points[i+1].Time.Unix() - points[i].Time.Unix()
			if seconds <= 0 {
				seconds = 1
			}
			waitPeriod := time.Duration(seconds) * time.Second
			time.Sleep(waitPeriod)
		}
		rmcList = append(rmcList, rmcStr)
		fmt.Println(rmcStr)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <path to file.gpx> <path to serialPort> [-b baudrate value]")
		os.Exit(0)
	}
	filePath := os.Args[1]
	// filePath = "example.gpx" //debug
	input, err := os.Open(filePath)
	if err != nil {
		fmt.Fprint(os.Stderr, "Open file error: %v\n", err)
		os.Exit(1)
	}
	defer input.Close()
	serialPortPath := os.Args[2]

	ports, err := serial.GetPortsList()
	if err != nil {
		fmt.Println("ERROR!", err.Error())
		os.Exit(1)
	}
	exists := false
	for _, port := range ports {
		if port == serialPortPath {
			exists = true
			break
		}
	}
	if !exists {
		fmt.Printf("%s port does not exist! \n", serialPortPath)
		os.Exit(1)
	}
	baudrate := flag.Int("b", 115200, "Baudrate")
	flag.CommandLine.Parse(os.Args[3:])
	fmt.Printf("Baudrate: %v", *baudrate)
	mode := &serial.Mode{
		BaudRate: *baudrate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(serialPortPath, mode)
	if err != nil {
		fmt.Printf("Serial port was not opened: %v\n", err)
		os.Exit(1)
	}
	if port == nil {
		fmt.Printf("Serial port was not opened: %v\n", err)
		os.Exit(1)
	}
	defer port.Close()

	var gpx gpxreader.GpxReader
	points, err := gpx.GetPoints(input)
	if err != nil {
		println("ERROR!", err.Error())
		os.Exit(1)
	}
	if err := keyboard.Open(); err != nil {
		fmt.Println(err.Error())
	}
	defer keyboard.Close()

	quit := make(chan struct{})

	go func() {
		for {
			char, _, err := keyboard.GetSingleKey()
			if err != nil {
				fmt.Println("Reading error:", err)
				continue
			}
			if char == 'q' || char == 'Q' {
				fmt.Println("Exit")
				close(quit)
				os.Exit(0)
			}
		}
	}()

	sendPointsToPort(points, port)
}
