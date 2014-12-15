// Package netint (linode-netint) is a client for accessing the Linode network
// internals samples. This API is undocumented and looks to be a set of
// unauthenticated endpoints that provide JSON data. This package also does
// some alterations to the data provided by Linode as some of the JSON types
// don't make sense...
//
// * The rount-trip-time (RTT) field is converted from a string to uint32
//
// * The Loss field is converted from a string to uint32
//
// * The Jitter field is converted from a string to a uint32
//
// To note, this package is not maintained by nor affiliated with Linode. It
// simply consumes data from an undocumented pulic API.
package netint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
)

// BASE_URLF is the base URL with a format specifier for the abbreviation
const BASE_URLF = "http://netint-%v.linode.com/ping/samples"

const (
	DALLAS  = "dal"   // Dallas abbreviation
	FREMONT = "fmt"   // Fremont abbreviation
	ATLANTA = "atl"   // Atlanta abbreviation
	NEWARK  = "nwk"   // Newark abbreviation
	LONDON  = "lon"   // London abbreviation
	TOKYO   = "tok"   // Tokyo abbreviation
	VERSION = "0.0.1" // Library version
)

// used for parsing the JSON response
type samples struct {
	Dallas  [][]interface{} `json:"linode-dallas"`
	Fremont [][]interface{} `json:"linode-fremont"`
	Atlanta [][]interface{} `json:"linode-atlanta"`
	Newark  [][]interface{} `json:"linode-newark"`
	London  [][]interface{} `json:"linode-london"`
	Tokyo   [][]interface{} `json:"linode-tokyo"`
}

// Sample is a single result for a point-to-point measurement.
type Sample struct {
	Epoch  int64
	RTT    uint32 // unit: milliseconds
	Loss   uint32 // unit: percentage
	Jitter uint32 // unit: milliseconds
}

// Overview is the entire view a single region has to the rest of the regions.
// It consists of one *Sample for each Region
type Overview struct {
	Name    string
	Dallas  *Sample
	Fremont *Sample
	Atlanta *Sample
	Newark  *Sample
	London  *Sample
	Tokyo   *Sample
}

// Regions is a function that returns a slice of strings that is the
// collection of Linode regions.
func Regions() []string {
	return []string{"dallas", "fremont", "atlanta", "newark", "london", "tokyo"}
}

// AllSamples is a function to return all overviews.
// It's a map of *Overview instances with the lowercase name
// of the region as the key.
func AllOverviews() (map[string]*Overview, error) {
	m := make(map[string]*Overview)

	// loop over each region and
	// populate its overview
	for _, d := range Regions() {
		o, err := getOverview(d)

		if err != nil {
			return nil, err
		}

		m[d] = o
	}

	return m, nil
}

// Dallas is a function to get an overview of the Dallas region.
func Dallas() (*Overview, error) {
	return getOverview("dallas")
}

// Fremont is a function to get an overview of the Fremont region.
func Fremont() (*Overview, error) {
	return getOverview("fremont")
}

// Atlanta is a function to get an overview of the Atlanta region.
func Atlanta() (*Overview, error) {
	return getOverview("atlanta")
}

// Newark is a function to get an overview of the Newark region.
func Newark() (*Overview, error) {
	return getOverview("newark")
}

// London is a function to get an overview of the London region.
func London() (*Overview, error) {
	return getOverview("london")
}

// Tokyo is a function to get an overview of the Tokyo region.
func Tokyo() (*Overview, error) {
	return getOverview("tokyo")
}

func getOverview(r string) (o *Overview, err error) {
	var u string

	// determine the URL based on the region
	// if the region is unknown return error
	switch r {
	case "dallas":
		u = url(DALLAS)
	case "fremont":
		u = url(FREMONT)
	case "atlanta":
		u = url(ATLANTA)
	case "newark":
		u = url(NEWARK)
	case "london":
		u = url(LONDON)
	case "tokyo":
		u = url(TOKYO)
	default:
		return nil, fmt.Errorf("'%v' is not a valid datacenter", r)
	}

	body, err := responseBody(u)

	if err != nil {
		return
	}

	s := &samples{}

	err = json.Unmarshal(body, s)

	if err != nil {
		return
	}

	o, err = buildOverview(s)

	if err != nil {
		return nil, err
	}

	o.Name = r

	return
}

func url(abbr string) string {
	return fmt.Sprintf(BASE_URLF, abbr)
}

func responseBody(url string) ([]byte, error) {
	httpc := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// we set a user agent so Linode has an idea of where requests are being generated from
	// LinodeNetInt/<VERSION> (go<VER> net/http)
	req.Header.Add("User-Agent", fmt.Sprintf("LinodeNetInt/%v (%v net/http)", VERSION, runtime.Version()))

	// execute the request
	resp, err := httpc.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// get the entire body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func buildOverview(s *samples) (o *Overview, err error) {
	o = &Overview{}

	o.Dallas, err = pullSample(s.Dallas)

	if err != nil {
		return nil, err
	}

	o.Fremont, err = pullSample(s.Fremont)

	if err != nil {
		return nil, err
	}

	o.Atlanta, err = pullSample(s.Atlanta)

	if err != nil {
		return nil, err
	}

	o.Newark, err = pullSample(s.Newark)

	if err != nil {
		return nil, err
	}

	o.London, err = pullSample(s.London)

	if err != nil {
		return nil, err
	}

	o.Tokyo, err = pullSample(s.Tokyo)

	if err != nil {
		return nil, err
	}

	return
}

func pullSample(i [][]interface{}) (s *Sample, err error) {
	// NOTE: As has been historically been a pain point with Linode,
	//       these endpoints provide some wonky JSON. Only the timestamp
	//       is in a useful format (numeric). RTT, Loss, and Jitter are all
	//       strings for some reason. So we need to get those values.

	// convert the RTT to a uint
	r, err := strconv.ParseUint(i[0][1].(string), 10, 32)

	if err != nil {
		return
	}

	// convert the Loss to a uint
	l, err := strconv.ParseUint(i[0][2].(string), 10, 32)

	if err != nil {
		return
	}

	// convert the jitter to a uint
	j, err := strconv.ParseUint(i[0][3].(string), 10, 32)

	if err != nil {
		return
	}

	s = &Sample{}

	// convert the UNIX timestamp to an int64
	s.Epoch = int64(i[0][0].(float64))

	s.RTT = uint32(r)
	s.Loss = uint32(l)
	s.Jitter = uint32(j)

	return
}
