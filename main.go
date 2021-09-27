package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"github.com/lensesio/tableprinter"
)

type OCPVersion struct {
	Version string `json:"version"`
}

type Version struct {
	Location   string       `header:"Location"`
	RPVersion  string       `header:"RPVersion"`
	OCPVersion []OCPVersion `header:"OCPVersions"`
}

func getOCPVersions(location string) ([]OCPVersion, error) {

	url := "https://arorpversion.blob.core.windows.net/ocpversions/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("failed to pull ocp versions: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read requests body: %s", err)
		return nil, err
	}

	var ocpVersions []OCPVersion
	err = json.Unmarshal(jsonData, &ocpVersions)
	if err != nil {
		log.Printf("failed to unmarshal ocp versions: %s", err)
		return nil, err
	}

	return ocpVersions, nil
}

func getRPVersions(location string) (string, error) {

	url := "https://arorpversion.blob.core.windows.net/rpversion/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Printf("failed to pull rp version: %s", err)
		return "", err
	}
	defer res.Body.Close()

	commitVersion, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read requests body: %s", err)
		return "", err
	}

	return string(commitVersion), nil
}

func processLocation(location string, ch chan Version) {

	ocpVersions, err := getOCPVersions(location)
	if err != nil {
		log.Printf("failed to get OCP versions for %q location: %s", location, err)
	}

	rpVersion, err := getRPVersions(location)
	if err != nil {
		log.Printf("failed to get RP version for %q location: %s", location, err)
	}

	locationResult := Version{
		Location:   location,
		RPVersion:  rpVersion,
		OCPVersion: ocpVersions,
	}

	ch <- locationResult
}

func main() {

	// TODO: Pull dynamically eventually
	locations := []string{
		"eastus2euap", "westcentralus", "australiaeast", "japaneast", "koreacentral",
		"australiasoutheast", "centralindia", "southindia", "japanwest", "eastasia",
		"centralus", "eastus", "eastus2", "northcentralus", "southcentralus",
		"westus", "westus2", "canadacentral", "canadaeast", "francecentral",
		"germanywestcentral", "northeurope", "norwayeast", "switzerlandnorth", "switzerlandwest",
		"westeurope", "brazilsouth", "brazilsoutheast", "southeastasia", "uaenorth",
		"southafricanorth", "uksouth", "ukwest",
	}

	locationsCount := len(locations)

	var wg sync.WaitGroup
	wg.Add(locationsCount)

	ch := make(chan Version, locationsCount)

	for _, location := range locations {

		location := location

		go func() {
			defer wg.Done()
			processLocation(location, ch)
		}()
	}

	wg.Wait()
	close(ch)

	var results []Version

	for result := range ch {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[j].Location > results[i].Location
	})

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
