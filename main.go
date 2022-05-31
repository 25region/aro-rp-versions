package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/25region/aro-rp-versions/pkg/version"
	"github.com/lensesio/tableprinter"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v3"
)

var log = logrus.New()

type flags struct {
	debug    bool
	location []string
	output   string
	version  bool
}

type OCPVersion struct {
	Version string `json:"version"`
}

type Result struct {
	Location    string   `header:"Location" json:"location"`
	RPVersion   string   `header:"RPVersion" json:"rpVersion"`
	OCPVersions []string `header:"OCPVersions" json:"ocpVersions"`
}

func getOCPVersions(location string) ([]string, error) {

	url := "https://arorpversion.blob.core.windows.net/ocpversions/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Debugf("failed to pull ocp versions for %q: %s", location, err)
		return nil, err
	}
	defer res.Body.Close()

	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debugf("failed to read requests body for %q: %s", location, err)
		return nil, err
	}

	var ocpVersions []OCPVersion
	err = json.Unmarshal(jsonData, &ocpVersions)
	if err != nil {
		log.Debugf("failed to unmarshal ocp versions for %q: %s", location, err)
		return nil, err
	}

	var result []string
	for _, version := range ocpVersions {
		result = append(result, fmt.Sprint(version.Version))
	}

	return result, nil
}

func getRPVersions(location string) (string, error) {

	url := "https://arorpversion.blob.core.windows.net/rpversion/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Debugf("failed to pull rp version for %q: %s", location, err)
		return "", err
	}
	defer res.Body.Close()

	commitVersion, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debugf("failed to read requests body for %q: %s", location, err)
		return "", err
	}

	return string(commitVersion), nil
}

func processLocation(location string, ch chan Result) {

	log.Debugf("processing location: %s", location)

	ocpVersions, err := getOCPVersions(location)
	if err != nil {
		log.Debugf("failed to get OCP versions for %q location: %s", location, err)
	}

	rpVersion, err := getRPVersions(location)
	if err != nil {
		log.Debugf("failed to get RP version for %q location: %s", location, err)
	}

	locationResult := Result{
		Location:    location,
		RPVersion:   rpVersion,
		OCPVersions: ocpVersions,
	}

	ch <- locationResult
}

func parseFlags() flags {

	var f flags
	flag.BoolVarP(&f.debug, "debug", "d", false, "enable debugging output")
	flag.StringSliceVarP(&f.location, "location", "l", []string{}, "comma-separated Azure regions")
	flag.StringVarP(&f.output, "output", "o", "table", "defines output format (table|json)")
	flag.BoolVarP(&f.version, "version", "v", false, "version")

	flag.Parse()

	return f
}

func main() {

	flags := parseFlags()

	// Configure logging level
	log.SetOutput(os.Stdout)
	if flags.debug {
		log.Level = logrus.DebugLevel
	}

	if flags.version {
		version.Print()
		os.Exit(0)
	}

	var locations []string
	if len(flags.location) > 0 {
		locations = flags.location
	} else {

		yamlFile, err := os.Open("locations.yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer yamlFile.Close()

		bytes, _ := ioutil.ReadAll(yamlFile)
		errYaml := yaml.Unmarshal(bytes, &locations)
		if errYaml != nil {
			fmt.Println(errYaml)
			os.Exit(1)
		}

	}

	locationsCount := len(locations)

	var wg sync.WaitGroup
	wg.Add(locationsCount)

	ch := make(chan Result, locationsCount)

	for _, location := range locations {

		location := location

		go func() {
			defer wg.Done()
			processLocation(location, ch)
		}()
	}

	wg.Wait()
	close(ch)

	var results []Result

	for result := range ch {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[j].Location > results[i].Location
	})

	switch o := flags.output; o {
	case "json":

		var err error
		in, err := json.Marshal(results)
		if err != nil {
			log.Fatalf("failed to marshal the result in json format: %s", err)
		}
		var out bytes.Buffer
		err = json.Indent(&out, []byte(in), "", "  ")
		if err != nil {
			log.Fatalf("failed to output the result in json format: %s", err)
		}
		fmt.Println(&out)

	default:
		printer := tableprinter.New(os.Stdout)

		// Optionally, customize the table, import of the underline 'tablewriter' package is required for that.
		// printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
		// printer.CenterSeparator = "│"
		// printer.ColumnSeparator = "│"
		// printer.RowSeparator = "─"
		// printer.HeaderBgColor = tablewriter.BgBlackColor
		// printer.HeaderFgColor = tablewriter.FgGreenColor

		// Print the slice of structs as table, as shown above.
		printer.Print(results)
	}

}
