package main

import (
  "fmt"
  "github.com/stratoberry/gps"
)

func main() {
  var dev *gps.Device
  var err error
  if dev, err = gps.Open("/dev/ttyAMA0"); err != nil {
    panic(err)
  }

  dev.Watch()

  for fix := range dev.Fixes {
    fmt.Println(fmt.Sprintf("Location: (%.8f,%.8f) Altitude:%.2fm Angle:%.2f", fix.Lat, fix.Lon, fix.Alt, fix.TrackAngle))
  }
}
