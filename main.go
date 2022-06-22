package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"vestaluna/tools"
	"vestaluna/wmts"

	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

var XML = []string{"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_128ppd_v04/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_Viking_MDIM21_ClrMosaic_global_232m/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_Shade_Global_256ppd_v06/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Moon/EQ/LRO_LOLA_ClrRoughness_Global_16ppd/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mercury/NP/Mercury_MESSENGER_mosaic_npole_250m_2013/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Titan/EQ/Titan_global_32ppd_ColorRatio_v2/1.0.0/WMTSCapabilities.xml",
	"https://trek.nasa.gov/tiles/Mars/EQ/Mars_MOLA_blend200ppx_HRSC_ClrShade_clon0dd_200mpp_lzw/1.0.0/WMTSCapabilities.xml"}

// xml file locations are stored in this dir in the apiEndPoints.txt file
func readApiEndpoints(filepath string) ([]string, error) {
	file, err := os.Open(filepath)

	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var xml []string
	for scanner.Scan() {
		xml = append(xml, scanner.Text())
		//	log.Println(xml)

	}

	return xml, nil

}

// Read in the .xml files to hit and populate the UI

type simpleCatalog struct {
	Catalog     string
	XMLLocation string
	Format      string
	LODs        int
	URL         string
}

// pull a simplified version of the catalog data for gui display
func pullSimpleCatalogData(XML []string) []simpleCatalog {
	var catalogEntries []simpleCatalog
	for _, xml := range XML {
		entry := wmts.LoadCatalog(xml)

		sc := simpleCatalog{
			Catalog:     entry.Contents.Layer.Identifier,
			XMLLocation: xml,
			Format:      entry.Contents.Layer.Format,
			LODs:        len(entry.Contents.TileMatrixSet.TileMatrix),
			URL:         entry.Contents.Layer.ResourceURL.Template,
		}
		catalogEntries = append(catalogEntries, sc)
	}

	return catalogEntries
}

func main() {

	a := app.New()

	var wg sync.WaitGroup

	w := a.NewWindow("vestaluna")

	w.Resize(fyne.NewSize(640, 360))

	xmlList := &XML

	sc := pullSimpleCatalogData(*xmlList)
	log.Println("LEN:", len(sc))
	for _, xml := range *xmlList {
		log.Println(xml)
	}

	listView := widget.NewList(func() int {
		return len(sc)

	}, func() fyne.CanvasObject {
		return widget.NewLabel("template")

	},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(sc[id].Catalog)
		})

	contentText := widget.NewLabel("Select a catalog")

	contentText.Wrapping = fyne.TextWrapWord

	lod := binding.NewString()
	ext := binding.NewString()
	catalogID := binding.NewString()

	listView.OnSelected = func(id widget.ListItemID) {
		extension := strings.Replace(sc[id].Format, "image/", "", 1)

		txt := fmt.Sprintf("Catalog:%s\nLODs:%d\nFormat:%s",
			sc[id].Catalog, sc[id].LODs, extension)
		contentText.Text = txt

		log.Println("Setting globals")
		ext.Set(extension)
		_, _ = ext.Get()
		log.Println("Set ext")

		catalogID.Set(sc[id].Catalog)
		_, _ = catalogID.Get()
		log.Println("Set cat")

		// HERE: this is where we're breaking things..
		var lodCurrent string = fmt.Sprintf("%d", (sc[id].LODs))
		lod.Set(lodCurrent)
		log.Println("Set lod")
		_, _ = lod.Get()

	}

	split := container.NewHSplit(
		listView,
		container.NewVBox(
			//TODO: Vsplit this for description/preview window/buttons
			//TODO: create a create image button -> spawn available resolutions dialouge
			//TODO: buttons needs to include the LOD setting
			//TODO: add a depth button/slider to allow us to change the LOD at runtime

			container.NewMax(contentText),
			widget.NewButton("Download", func() {
				log.Println("Downloading")
				//TODO: how to get the right index down here after it's set up there..
				wg.Add(1)
				go func(wg *sync.WaitGroup, idx int) {
					defer wg.Done()
					lodCurrent, _ := lod.Get()
					lod, _ := strconv.Atoi(lodCurrent)
					if wmts.FetchExact(sc[idx].XMLLocation, lod) {
						log.Println("Download Complete")
					} else {
						log.Println("Download was incomplete...")
						wmts.FetchExact(sc[idx].XMLLocation, lod)
					}

				}(&wg, 1) //NOTE: explicitly passing in mars atm.
			}),
			widget.NewButton("Concat", func() {
				log.Println("Concatenating")
				catalog, _ := catalogID.Get()
				dirpath := filepath.Join("downloads", catalog)
				ext, _ := ext.Get()
				lodCurrent, _ := lod.Get()
				lod, _ := strconv.Atoi(lodCurrent)
				tools.ConcatWithPython(dirpath, lod, ext)
				log.Println("Concatenation Complete")

			}),
			// run the xdg-open command to open the folder
			widget.NewButton("Disk", func() {
				log.Println("Opening disk")
				cmd := exec.Command("xdg-open", "downloads")
				err := cmd.Run()
				if err != nil {
					log.Println(err)
				}

			}),
		),
	)

	w.SetContent(split)

	w.ShowAndRun()

}