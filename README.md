# aro-rp-versions

## Description
Pulls ARO RP versions and their corresponding upgrade streams for all regions in a table format by default or json if requested via the output argument.
This uses go routines which makes pulling from all endpoints almost instantaneous.

## Usage:
```
./aro-rp-versions <args>
  -d, --debug              enable debugging output
  -l, --location strings   comma-separated Azure regions
  -o, --output string      defines output format (table|json) (defaults to "table")
```

## Examples:
- Default settings: all regions, output as a table
```bash
./aro-rp-versions
```

- Specify only some regions to be included and output as json
```bash
./aro-rp-versions -l westus,westus2 -o json
```

- Debug output to output error:
```bash
./aro-rp-versions -d
```
