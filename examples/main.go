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
    fmt.Println(fix)
  }
}
