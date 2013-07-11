# gps

*Neatly packaged GPS parser for Go*

## Why on earth?

Speed! Memory usage! Reasons!

The existing set up with [GPSd](http://gpsd.berlios.de), was a bit too hefty for our use case. GPSd would alone use around 5MB of our precious, but rather limited, 256MB Raspberry Pi + another 2MB for our [go-gpsd](http://github.com/stratoberry/go-gpsd) client.

gps, on the other hand, is tuned for performance and minimal memory consumption. A simple application based on gps shouldn't use more than 2MB of RAM and no more than 1.3% of CPU* - since it doesn't have any of GPSd's fancy features like hot-plugging, JSON interface or support for various devices. gps simply listens for [GPGGA](http://aprs.gids.nl/nmea/#gga) and [GPRMC](http://aprs.gids.nl/nmea/#rmc) sentences (and no other), parses them, combines into a single fix and hands them off to the user.

&#42; Disclamer: all measurements were made on a first version, model B Raspberry Pi with 256 megs of RAM.

## Installation


<pre><code># go get github.com/stratoberry/gps</code></pre>


## Usage

    dev, err := gps.Open("/dev/ttyAMA0")
    if err != nil {
        panic(err)
    }
        
    dev.Watch()
    for fix := range dev.Fixes {
        fmt.Println(fix.Lat, fix.Lon, fix.Alt)
    }

`gps.Device` has a `Fixes` channel (`chan *gps.Fix`) to which all GPS fixes will be sent.

## Documentation

Inspect the `gps.go` file in this repository or take a look at [godocs](http://godoc.org/github.com/stratoberry/gps]).

To find out more about the Stratoberry Pi project, take a look at our web site: [stratoberry.foi.hr](http://stratoberry.foi.hr)

## License

gps is available freely under the MIT license, for more details take a look at the `LICENSE` file within repository.