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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/widget"
)

func Preview() {
	myApp := app.New()
	w := myApp.NewWindow("Image")
	w.Resize(fyne.NewSize(960, 420))

	image := canvas.NewImageFromFile("stitched_results/1_Mars_Viking_MDIM21_ClrMosaic_global_232m.jpg")
	w.SetContent(image)

	w.ShowAndRun()
}

func Fetch1sOnStartup(sc []wmts.SimpleCatalog, wg *sync.WaitGroup) bool {
	defer wg.Done()
	fetchWg := sync.WaitGroup{}
	for idx := 1; idx < len(sc); idx++ {
		dirpath := filepath.Join("stitched_results", "2_"+sc[idx].Catalog+".jpg") //NOTE: Explicity using jpeg for all textures atm in the GUI version of the tool
		if !wmts.IsAlreadyDownloaded(dirpath) {
			fetchWg.Add(1)
			go func() {
				defer fetchWg.Done()
				log.Println("Fetching preview for: ", dirpath)
				if !wmts.FetchExact(sc[idx].XMLLocation, 2) {
					log.Println("Failed to get:", dirpath)
				}
				log.Println("Fetched preview")

			}()
		}
		log.Println(sc[idx].XMLLocation, "Is ok disk.")
	}

	//NOTE DON'T GOROUTINE THE PYTHON CALLS!
	fetchWg.Wait()
	for idx := 0; idx < len(sc); idx++ {
		dirpath := filepath.Join("downloads", sc[idx].Catalog)
		tools.ConcatWithPython(dirpath, 2)
	}

	return true
}

func main() {

	var wg sync.WaitGroup

	// Pull and serve (simfle) data for the UI
	xmlList, _ := tools.ReadApiEndpoints("apiEndPoints.txt")
	sc := wmts.PullSimpleCatalogData(xmlList)
	wg.Add(1)
	go Fetch1sOnStartup(sc, &wg) //Q: is this gonna clone it for me?
	wg.Wait()

	// UI bindings (fancy fyne mutex globals)
	lodSelect := binding.NewInt()
	ext := binding.NewString()
	catalogID := binding.NewString()
	catalogName := binding.NewString()

	// The GUI
	a := app.New()
	w := a.NewWindow("vestaluna")
	w.Resize(fyne.NewSize(960, 420))

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
			lodSelect.Set(maxLOD)

		}
		lodSelect.Set(parsedValue)
	})
	combo.SetSelectedIndex(0)

	split := container.NewHSplit(

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

					lod, _ := lodSelect.Get()

					tools.ConcatWithPython(dirpath, lod-1) // Again, the LOD is set by the UI which is NOT 0 indexed
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

					lod, _ := lodSelect.Get()
					idx, _ := strconv.Atoi(catID)
					result := strconv.Itoa(lod-1) + "_" + sc[idx].Catalog + ".jpg"
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
			layout.NewSpacer(),
			pbar,
		),
	)

	w.SetContent(split)

	w.ShowAndRun()
	wg.Wait()

}
