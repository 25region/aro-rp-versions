package ocp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/25region/aro-rp-versions/pkg/logger"
)

type OCPVersion struct {
	Version string `json:"version"`
}

func GetOCPVersions(location string) ([]string, error) {

	url := "https://arorpversion.blob.core.windows.net/ocpversions/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		logger.Log.Debugf("failed to pull ocp versions for %q: %s", location, err)
		return nil, err
	}
	defer res.Body.Close()

	jsonData, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Log.Debugf("failed to read requests body for %q: %s", location, err)
		return nil, err
	}

	var ocpVersions []OCPVersion
	err = json.Unmarshal(jsonData, &ocpVersions)
	if err != nil {
		logger.Log.Debugf("failed to unmarshal ocp versions for %q: %s", location, err)
		return nil, err
	}

	var result []string
	for _, version := range ocpVersions {
		result = append(result, fmt.Sprint(version.Version))
	}

	return result, nil
}

func GetRPVersions(location string) (string, error) {

	url := "https://arorpversion.blob.core.windows.net/rpversion/" + location

	res, err := http.Get(url)

	if err != nil || res.StatusCode != http.StatusOK {
		logger.Log.Debugf("failed to pull rp version for %q: %s", location, err)
		return "", err
	}
	defer res.Body.Close()

	commitVersion, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Log.Debugf("failed to read requests body for %q: %s", location, err)
		return "", err
	}

	return string(commitVersion), nil
}
