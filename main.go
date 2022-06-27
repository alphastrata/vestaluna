//
// VESTALUNA
// a NASA database scraper and concatenation tool for equirectangular MASSIVELY detailed datasets.
//

//TODO: Vsplit this for description/preview window/buttons
//TODO: The xml read from disk is broken >< stop using the GLOBAL!
//TODO: create a create image button -> spawn available resolutions dialouge
//TODO: buttons needs to include the LOD setting
//TODO: add a depth button/slider to allow us to change the LOD at runtime
//TODO: progressbars are nice in the cli-tool, how about bringing them into the UI?
package main

import (
	"fmt"
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
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/widget"
)

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

// Updates UI progress bar's value by `v`
func updatePB(pb *widget.ProgressBar, catalogID string, lod int, pbMax float64) {
	// get v by counting the files relative to the download query on a sep thread..
	for {
		v := float64(0.0) //NOTE: needs to work out its value by counting the files downloaded
		if v >= pbMax {
			return
		}
		pb.SetValue(v)

	}
}
func main() {

	a := app.New()

	var wg sync.WaitGroup
	//catIDX := binding.NewInt()
	lodSelect := binding.NewInt()
	ext := binding.NewString()
	catalogID := binding.NewString()
	catalogName := binding.NewString()

	w := a.NewWindow("vestaluna")
	w.Resize(fyne.NewSize(960, 420))

	xmlList, _ := tools.ReadApiEndpoints("apiEndPoints.txt")

	sc := pullSimpleCatalogData(xmlList)

	listView := widget.NewList(func() int {
		return len(sc)

	}, func() fyne.CanvasObject {
		return widget.NewLabel("template")

	},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(sc[id].Catalog)
			catalogName.Set(sc[id].Catalog)
		})

	ctTextBind := binding.NewString()
	ctTextBind.Set("Select a catalog")
	contentText := widget.NewLabelWithData(ctTextBind)

	contentText.Wrapping = fyne.TextWrapWord

	listView.OnSelected = func(id widget.ListItemID) {
		extension := strings.Replace(sc[id].Format, "image/", "", 1)

		ctTextBind.Set(fmt.Sprintf("Catalog:%s\nLODs:%d\nFormat:%s", sc[id].Catalog, sc[id].LODs, extension))

		ext.Set(extension)
		catalogID.Set(strconv.Itoa(id))
		lodSelect.Set(sc[id].LODs)

	}

	pbar := widget.NewProgressBarInfinite()
	pbar.Hide()

	combo := widget.NewSelect([]string{"LOD1", "LOD2", "LOD3", "LOD4", "LOD5", "LOD6", "LOD7", "LOD8"}, func(value string) {
		parsedValue, err := strconv.Atoi(strings.ReplaceAll(value, "LOD", ""))
		if err != nil {
			log.Fatal(err)
		}

		maxLOD, _ := lodSelect.Get()
		if parsedValue > maxLOD {
			lodSelect.Set(parsedValue)

		}
		lodSelect.Set(maxLOD)
	})
	combo.SetSelectedIndex(0)

	split := container.NewHSplit(

		listView,
		container.NewVBox(

			container.NewMax(contentText),
			widget.NewButton("Download", func() {
				//TODO: how to get the right index down here after it's set up there..
				catIDCurrent, err := catalogID.Get()
				if err != nil {
					log.Fatal("Error catalogID.Get()", err)
				}

				catID, err := strconv.Atoi(catIDCurrent)
				if err != nil {
					log.Fatal("Error parsing catIDCurrent into int -- maybe it received invalid data", err)
				}

				lod, err := lodSelect.Get()
				if err != nil {
					log.Fatal("Error lod.Get()", err)
				}
				pbar.Show()

				wg.Add(1)
				go func(wg *sync.WaitGroup, idx int, lod int) {
					defer wg.Done()

					log.Println("Downloading...")
					if wmts.FetchExact(sc[idx].XMLLocation, lod) {
						log.Println("Download Complete")
						pbar.Hide()
					} else {
						log.Println("Download was incomplete...")
						wmts.FetchExact(sc[idx].XMLLocation, lod)
					}

				}(&wg, catID, lod)
				wg.Wait()
			}),

			widget.NewButton("Concat", func() {
				log.Println("Concatenating")
				pbar.Show()
				catID, _ := catalogID.Get()
				idx, _ := strconv.Atoi(catID)
				dirpath := filepath.Join("downloads", sc[idx].Catalog)

				lod, _ := lodSelect.Get()

				tools.ConcatWithPython(dirpath, lod)
				log.Println("Concatenation Complete")
				pbar.Hide()

			}),
			widget.NewButton("Tiles", func() {
				catID, _ := catalogID.Get()
				idx, _ := strconv.Atoi(catID)
				dirpath := filepath.Join("downloads", sc[idx].Catalog)

				log.Println("Opening disk")
				cmd := exec.Command("xdg-open", dirpath)
				err := cmd.Run()
				if err != nil {
					log.Println(err)
				}

			}),
			widget.NewButton("ConcatResults", func() {
				catID, _ := catalogID.Get()

				lod, _ := lodSelect.Get()
				idx, _ := strconv.Atoi(catID)
				result := strconv.Itoa(lod) + "_" + sc[idx].Catalog + ".jpg"
				concatPath := filepath.Join("stitched_results", result)

				cmd := exec.Command("xdg-open", concatPath)
				err := cmd.Run()
				if err != nil {
					log.Println(err)
				}

			}),
			combo,
			layout.NewSpacer(),
			pbar,
		),
	)

	w.SetContent(split)

	w.ShowAndRun()
	wg.Wait()

}
