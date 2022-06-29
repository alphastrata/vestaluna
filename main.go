//
// VESTALUNA
// a NASA database scraper and concatenation tool for equirectangular MASSIVELY detailed datasets.
//

//TODO: impl a preview window of LOD1, for any selected catalog

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

	"fyne.io/fyne/v2/widget"
)

//func Preview() {
//	myApp := app.New()
//	w := myApp.NewWindow("Image")
//	w.Resize(fyne.NewSize(960, 420))
//
//	image := canvas.NewImageFromFile("stitched_results/1_Mars_Viking_MDIM21_ClrMosaic_global_232m.jpg")
//	w.SetContent(image)
//
//	w.ShowAndRun()
//}

func main() {
	tools.InitDirStructure()

	var wg sync.WaitGroup

	// Pull and serve (simfle) data for the UI
	xmlList, _ := tools.ReadApiEndpoints("apiEndPoints.txt")
	sc := wmts.PullSimpleCatalogData(xmlList)

	// UI bindings (fancy fyne mutex globals)
	lodSelect := binding.NewInt()
	lodMax := binding.NewInt()
	ext := binding.NewString()
	catalogID := binding.NewString()
	catalogName := binding.NewString()

	// The GUI
	a := app.New()
	w := a.NewWindow("vestaluna")
	w.Resize(fyne.NewSize(1200, 600))

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
		lodSelect.Set(1) // NOTE: start the lod at 0
		lodMax.Set(sc[id].LODs)

	}

	pbar := widget.NewProgressBarInfinite()
	pbar.Hide()

	combo := widget.NewSelect([]string{"LOD1", "LOD2", "LOD3", "LOD4", "LOD5", "LOD6", "LOD7", "LOD8", "LOD9"}, func(value string) {
		parsedValue, err := strconv.Atoi(strings.ReplaceAll(value, "LOD", ""))
		if err != nil {
			log.Fatal(err)
		}

		maxLOD, _ := lodMax.Get()
		if parsedValue > maxLOD {
			lodSelect.Set(maxLOD)

		} else {
			lodSelect.Set(parsedValue)
		}
	})
	combo.SetSelectedIndex(0)

	split := container.NewVBox(
		container.NewHSplit(

			listView,
			container.NewVBox(

				container.NewMax(contentText),

				// Get tiles a user has requested
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

					pbar.Show()

					wg.Add(1)
					go func(wg *sync.WaitGroup, idx int) {
						defer wg.Done()

						lod, err := lodSelect.Get()
						if err != nil {
							log.Fatal("Error lod.Get()", err)
						}
						lod = lod - 1 // the LOD is set by the UI which is NOT 0 indexed

						log.Println("Downloading...")
						if wmts.FetchExact(sc[idx].XMLLocation, lod) {
							log.Println("Download Complete")
							pbar.Hide()
						} else {
							log.Println("Download was incomplete...")
							wmts.FetchExact(sc[idx].XMLLocation, lod)
						}

					}(&wg, catID)
					wg.Wait()
				}),

				widget.NewButton("Concat", func() {
					wg.Add(1)
					go func() {
						defer wg.Done()
						log.Println("Concatenating")
						pbar.Show()
						catID, _ := catalogID.Get()
						idx, _ := strconv.Atoi(catID)
						dirpath := filepath.Join("downloads", sc[idx].Catalog)

						lod, err := lodSelect.Get()
						if err != nil {
							log.Fatal("Error lod.Get()", err)
						}
						lod = lod - 1 // the LOD is set by the UI which is NOT 0 indexed

						tools.ConcatWithPython(dirpath, lod)
						log.Println("Concatenation Complete")
						pbar.Hide()

					}()
				}),
				widget.NewButton("Tiles", func() {
					wg.Add(1)
					go func() {
						defer wg.Done()
						catID, _ := catalogID.Get()
						idx, _ := strconv.Atoi(catID)
						dirpath := filepath.Join("downloads", sc[idx].Catalog)

						log.Println("Opening disk")
						cmd := exec.Command("xdg-open", dirpath)
						output, err := cmd.Output()
						if err != nil {
							log.Println(err)
							log.Println(output)
						}
					}()
				}),
				widget.NewButton("ConcatResults", func() {
					wg.Add(1)
					go func() {
						defer wg.Done()
						catID, _ := catalogID.Get()

						lod, err := lodSelect.Get()
						if err != nil {
							log.Fatal("Error lod.Get()", err)
						}
						lod = lod - 1 // the LOD is set by the UI which is NOT 0 indexed

						idx, _ := strconv.Atoi(catID)
						result := strconv.Itoa(lod) + "_" + sc[idx].Catalog + ".jpg"
						concatPath := filepath.Join("stitched_results", result)

						cmd := exec.Command("xdg-open", concatPath)
						output, err := cmd.Output()
						if err != nil {
							log.Println(err)
							log.Println(output)
						}

					}()
				}),
				combo,
				pbar,
			),
		),
	)
	w.SetContent(split)

	w.ShowAndRun()
	wg.Wait()

}
