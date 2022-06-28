package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"vestaluna/tools"
	"vestaluna/wmts"
)

type Args struct {
	mode string //concat or fetch
	xml  string //Use when fetching
	lod  string //LOD to scrape until, or concat at
	dir  string //directory to store files in
}

func parseArgs() Args {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage: ./main.go <mode> <lod> <dir>")
		fmt.Println("mode: fetch or concat")
		fmt.Println("xml: there, many options, copy and paste one of the below")
		fmt.Println("lod: LOD to scrape until, or concat at")
		fmt.Println("dir: directory to store files in")
		fmt.Println("Databases I have information for:")
		for _, xml := range XML {
			fmt.Println(xml)
		}

		fmt.Println("edit the main.go file's idx to get the database catalog you want")
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Println("Usage: go run main.go fetch")
		fmt.Println("Usage: go run main.go concat")
		os.Exit(1)
	}

	if args[1] == "fetch" {
		return Args{mode: args[1],
			lod: args[3],
			dir: "",
			xml: args[2]}
	}

	return Args{mode: args[1],
		lod: args[3],
		dir: args[4],
		xml: args[2],
	}
}

var XML = []string{"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_128ppd_v04/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_Viking_MDIM21_ClrMosaic_global_232m/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_256ppd_v06/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_ClrRoughness_Global_16ppd/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mercury/NP/Mercury_MESSENGER_mosaic_npole_250m_2013/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Titan/EQ/Titan_global_32ppd_ColorRatio_v2/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_MOLA_blend200ppx_HRSC_ClrShade_clon0dd_200mpp_lzw/1.0.0/WMTSCapabilities.xml"}

func main() {
	args := parseArgs()
	var wmtsXML string

	for _, s := range XML {
		if strings.Contains(s, args.dir) {
			wmtsXML = s
		}

	}

	switch args.mode {
	case "fetch":
		// remove the misses.txt
		os.Remove("misses.txt")
		lod, _ := strconv.Atoi(args.lod)
		wmts.FetchExact(wmtsXML, lod)
	case "concat":
		sp := strings.Split(wmtsXML, "/")
		dirpath := "downloads/" + sp[len(sp)-2]
		lod, _ := strconv.Atoi(args.lod)
		tools.ConcatWithPython(dirpath, lod)
	}
}
