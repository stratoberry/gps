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

  var n int = 1
  for fix := range dev.Fixes {
    fmt.Println(fmt.Sprintf("%d Location: (%.5f,%.5f) Altitude:%.2fm", n, fix.Lat, fix.Lon, fix.Alt))
    n++
    if n == 10 {
      dev.Close()
    }
  }
}
