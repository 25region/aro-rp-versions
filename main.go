package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/25region/aro-rp-versions/pkg/logger"
	"github.com/25region/aro-rp-versions/pkg/ocp"
	"github.com/25region/aro-rp-versions/pkg/version"
	"github.com/lensesio/tableprinter"
	"github.com/sirupsen/logrus"

	flag "github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v3"
)

const locationsFilePath = "locations.yaml"

type Result struct {
	Location    string   `header:"Location" json:"location"`
	RPVersion   string   `header:"RPVersion" json:"rpVersion"`
	OCPVersions []string `header:"OCPVersions" json:"ocpVersions"`
}

//go:embed locations.yaml
var f embed.FS

type flags struct {
	debug    bool
	location []string
	output   string
	version  bool
}

func getLocationsFromFile(path string) []string {

	var locations = []string{}

	yamlFile, err := f.Open(path)
	if err != nil {
		logger.Log.Fatalf("error: %v\n", err)
	}
	defer yamlFile.Close()

	bytes, _ := io.ReadAll(yamlFile)
	if err != nil {
		logger.Log.Fatalf("error: %v\n", err)
	}
	errYaml := yaml.Unmarshal(bytes, &locations)
	if errYaml != nil {
		logger.Log.Fatalf("error: %v\n", errYaml)
	}

	return locations
}

func processLocation(location string, ch chan Result) {

	logger.Log.Debugf("processing location: %s", location)

	ocpVersions, err := ocp.GetOCPVersions(location)
	if err != nil {
		logger.Log.Debugf("failed to get OCP versions for %q location: %s", location, err)
	}

	rpVersion, err := ocp.GetRPVersions(location)
	if err != nil {
		logger.Log.Debugf("failed to get RP version for %q location: %s", location, err)
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

	// Configure loggerging level
	logger.Log.SetOutput(os.Stdout)
	if flags.debug {
		logger.Log.Level = logrus.DebugLevel
	}

	if flags.version {
		version.Print()
		os.Exit(0)
	}

	var locations []string
	if len(flags.location) > 0 {
		locations = flags.location
	} else {
		locations = getLocationsFromFile(locationsFilePath)
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
			logger.Log.Fatalf("failed to marshal the result in json format: %s", err)
		}
		var out bytes.Buffer
		err = json.Indent(&out, []byte(in), "", "  ")
		if err != nil {
			logger.Log.Fatalf("failed to output the result in json format: %s", err)
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
