package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"vestaluna/tools"
	"vestaluna/wmts"

	"fyne.io/fyne/v2/widget"
)

type Args struct {
	mode string //concat or fetch
	lod  string //LOD to scrape until, or concat at
	dir  string //directory to store files in
}

func parseArgs() Args {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage: ./main.go <mode> <lod> <dir>")
		fmt.Println("mode: fetch or concat")
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
			lod: args[2],
			dir: ""} //N/A
	}

	return Args{mode: args[1],
		lod: args[2],
		dir: args[3]} // Location of dataset from which to concat
}

var XML = []string{"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_128ppd_v04/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_Viking_MDIM21_ClrMosaic_global_232m/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_256ppd_v06/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_ClrRoughness_Global_16ppd/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mercury/NP/Mercury_MESSENGER_mosaic_npole_250m_2013/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Titan/EQ/Titan_global_32ppd_ColorRatio_v2/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_MOLA_blend200ppx_HRSC_ClrShade_clon0dd_200mpp_lzw/1.0.0/WMTSCapabilities.xml"}

func main() {
	wmtsXML := XML[3]
	args := parseArgs()
	switch args.mode {
	case "fetch":
		// remove the misses.txt
		os.Remove("misses.txt")
		uiPbar := widget.NewProgressBar() //NOTE: Never used just passing as a dummy, may need to address later..
		lod, _ := strconv.Atoi(args.lod)
		wmts.FetchExact(wmtsXML, lod, uiPbar)
	case "concat":
		sp := strings.Split(wmtsXML, "/")
		dirpath := "downloads/" + sp[len(sp)-2]
		log.Println("CLI>>> Dirpath is:", dirpath)
		lod, _ := strconv.Atoi(args.lod)
		tools.ConcatWithPython(dirpath, lod)
	}
}
