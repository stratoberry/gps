package gps

import (
  "bufio"
  _ "fmt"
  "math"
  "os"
  "runtime"
  "strconv"
  "strings"
)

type Device struct {
  file         *os.File
  reader       *bufio.Reader
  open         bool
  Fix, nextFix *GPSFix

  // Fixes will receive new GPS fixes. This channel is unbuffered.
  Fixes chan *GPSFix
}

type GPSFix struct {
  Time                             int
  Quality                          int
  Satellites                       int
  Lat, Lon, Alt, Speed, TrackAngle float64
}

func (g *GPSFix) isComplete() bool {
  return g.Quality != 0 && g.Satellites != 0 && g.Lat != 0.0 && g.Lon != 0.0 && g.Alt != 0.0
}

// Open tries to access the specified device path
func Open(path string) (dev *Device, err error) {
  dev = &Device{}
  dev.file, err = os.Open(path)
  if err != nil {
    return
  }

  dev.reader = bufio.NewReader(dev.file)
  dev.open = true

  dev.Fix = &GPSFix{}
  dev.nextFix = &GPSFix{}
  dev.Fixes = make(chan *GPSFix)
  return
}

// Watch will spawn a new goroutine in which it will read and process GPS data from the device.
// Processed data will be available on the Fixes channel.
//
// Example
//    dev, _ := gps.Open("/dev/ttyAMA0")
//    dev.Watch()
//    for fix := range dev.Fixes {
//      fmt.Println(fix.Lat, fix.Lon)
//    }
//
func (d *Device) Watch() {
  go watchDevice(d)
}

// Close stops data parsing. Once closed, a Device can't be opened again,
// instead you will have to instantiate a new one.
func (d *Device) Close() {
  d.open = false
}

// FixTokens contain field names of specific NMEA sentences.
// Newly added tokens are currently ignored.
var FixTokens = map[string][]string{
  "$GPGGA": []string{"_", "Time", "Lat", "Lat", "Lon", "Lon", "Quality", "Satellites", "_", "Alt"},
  "$GPRMC": []string{"_", "Time", "Status", "Lat", "Lat", "Lon", "Lon", "Speed", "Angle"},
}

func watchDevice(d *Device) {
  var n int

  for {
    if d.open == false {
      d.file.Close()
      close(d.Fixes)
      return
    }

    n++
    if n%200 == 0 {
      n = 1
      runtime.GC()
    }

    if line, err := d.reader.ReadString('\n'); err == nil {
      isGGA := strings.HasPrefix(line, "$GPGGA")
      isRMC := strings.HasPrefix(line, "$GPRMC")
      if !isGGA && !isRMC {
        continue
      }

      tokenized := tokenizeString(strings.TrimSpace(line))

      t := parseInt(tokenized["Time"])
      if t < d.nextFix.Time {
        continue
      } else {
        d.nextFix.Time = t
      }

      d.nextFix.Lat = parseLatLon(tokenized["Lat"])
      d.nextFix.Lon = parseLatLon(tokenized["Lon"])

      if isGGA {
        d.nextFix.Quality = parseInt(tokenized["Quality"])
        d.nextFix.Satellites = parseInt(tokenized["Satellites"])
        d.nextFix.Alt = parseFloat(tokenized["Alt"])
      }

      if isRMC {
        d.nextFix.Speed = parseFloat(tokenized["Speed"])
        d.nextFix.TrackAngle = parseFloat(tokenized["Angle"])
      }

      if d.nextFix.isComplete() {
        d.Fixes <- d.nextFix
        d.Fix = d.nextFix
        d.nextFix = &GPSFix{}
      }
    }
  }
}

func tokenizeString(s string) map[string][]string {
  parts := strings.Split(s, ",")
  result := make(map[string][]string)
  tokens, ok := FixTokens[parts[0]]
  if ok == false {
    return result
  }

  for i, token := range tokens {
    if token != "_" {
      result[token] = append(result[token], parts[i])
    }
  }

  return result
}

func parseLatLon(fields []string) (decimal float64) {
  var err error
  var msf float64
  var deg int

  dms := strings.Split(fields[0], ".")
  dm := dms[0]
  d := dm[:len(dm)-2]
  if deg, err = strconv.Atoi(d); err != nil {
    return
  }

  m := dm[len(d):]
  ms := m + "." + dms[1]
  if msf, err = strconv.ParseFloat(ms, 64); err != nil {
    return
  }

  decimal = round(float64(deg)+msf/60.0, 8)

  if fields[1] == "W" || fields[1] == "S" {
    decimal *= -1
  }

  return
}

func parseInt(fields []string) (value int) {
  value, _ = strconv.Atoi(fields[0])
  return
}

func parseFloat(fields []string) (value float64) {
  value, _ = strconv.ParseFloat(fields[0], 64)
  return
}

// copied from gophers mailing list
func round(x float64, prec int) float64 {
  var rounder float64
  pow := math.Pow(10, float64(prec))
  intermed := x * pow

  if intermed < 0.0 {
    intermed -= 0.5
  } else {
    intermed += 0.5
  }
  rounder = float64(int64(intermed))

  return rounder / float64(pow)
}
