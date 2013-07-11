package gps

import (
  "bufio"
  "math"
  "os"
  "runtime"
  "strconv"
  "strings"
)

type Device struct {
  file   *os.File
  reader *bufio.Reader
  open   bool

  // Currently active GPSFix
  Fix     *GPSFix
  nextFix *GPSFix

  // Fixes will receive new GPS fixes. This channel is unbuffered.
  Fixes chan *GPSFix
}

type GPSFix struct {
  // Timestamp at which the fix was generated, format HHMMSS
  Time int

  // Quality of GPS fix, 0 - invalid, 1 - GPS fix, 2 - DGPS fix (less accurate)
  Quality int

  // Number of satelltes in view
  Satellites int

  // Latitude, longitude
  Lat, Lon float64

  // Altitude in meters
  Alt float64

  // Speed over ground in knots
  Speed float64

  // Magnetic variation
  TrackAngle float64
}

// Compares two GPS fixes for equality.
// Speed, Quality and Satellites fields are not checked.
func (g *GPSFix) Equals(b *GPSFix) bool {
  return g.Lat == b.Lat && g.Lon == b.Lon && g.Alt == b.Alt && g.TrackAngle == b.TrackAngle
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
  "$GPRMC": []string{"_", "Time", "Status", "Lat", "Lat", "Lon", "Lon", "Speed", "_", "_", "TrackAngle", "TrackAngle"},
}

func watchDevice(d *Device) {
  var n int
  var hasGGA = false
  var hasRMC = false

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
        //hasGGA = true
        d.nextFix.Quality = parseInt(tokenized["Quality"])
        d.nextFix.Satellites = parseInt(tokenized["Satellites"])
        alt := parseFloat(tokenized["Alt"])
        if alt < d.nextFix.Alt-0.1 || alt > d.nextFix.Alt+0.1 {
          d.nextFix.Alt = alt
          hasGGA = true
        }
      }

      if isRMC {
        hasRMC = true
        d.nextFix.Speed = parseFloat(tokenized["Speed"])
        d.nextFix.TrackAngle = parseLatLon(tokenized["TrackAngle"])
      }

      if hasGGA && hasRMC && !d.nextFix.Equals(d.Fix) {
        d.Fixes <- d.nextFix
        d.Fix = d.nextFix
        d.nextFix = &GPSFix{}
        hasGGA = false
        hasRMC = false
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

  if len(fields) != 2 || len(fields[0]) == 0 {
    return
  }

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
