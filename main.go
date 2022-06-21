package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"vestaluna/tools"
	"vestaluna/wmts"

	log "github.com/sirupsen/logrus"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
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

type guiVals struct {
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

	listView.OnSelected = func(id widget.ListItemID) {
		txt := fmt.Sprintf("Catalog:%s\nLODs:%d\nFormat:%s",
			sc[id].Catalog, sc[id].LODs, strings.Replace(sc[id].Format, "image/", "", 1))
		contentText.Text = txt
		//TODO: display preview of the lowest LOD available (tile 0/0/0.png)

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
				go func(wg *sync.WaitGroup) {
					wg.Add(1)
					defer wg.Done()
					if wmts.FetchExact(sc[5].XMLLocation, 3) {
						log.Println("Download Complete")
					} else {
						log.Println("Download was incomplete...")
						wmts.FetchExact(sc[5].XMLLocation, 3)
					}

				}(&wg)
			}),
			widget.NewButton("Concat", func() {
				log.Println("Concatenating")
				dirpath := filepath.Join("downloads", sc[2].Catalog)
				callConcat(dirpath, "3", true)
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

func callConcat(catalog, LOD string, show bool) {
	log.Println("Concatenating")
	tools.RunConcatenations(LOD, catalog)

}
