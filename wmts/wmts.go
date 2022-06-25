package wmts

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"sync"

	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/schollz/progressbar/v3"
)

// Check erro helper ( because I disklike peppering my code with if err != nil checks )
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

// Build a valid url from the parameters set out in the .xml spec
func buildURL(url string, style string, tileMatrixSet string, zoom *int, row int, col int) string {
	url = strings.Replace(url, "/{Style}", style, -1) // There is always a duplicate / in this one :(
	url = strings.Replace(url, "{TileMatrixSet}", tileMatrixSet, -1)
	url = strings.Replace(url, "{TileMatrix}", strconv.Itoa(*zoom), -1)
	url = strings.Replace(url, "{TileRow}", strconv.Itoa(row), -1)
	url = strings.Replace(url, "{TileCol}", strconv.Itoa(col), -1)

	return url

}

//check if the file is already downloaded
func isAlreadyDownloaded(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

//make a filename from the url and catalog
func makeFilenameFromURL(url string, catalog string) string {

	sp := strings.Split(url, "default028mm/")[1]
	sp = strings.Replace(sp, "/", "_", 3)

	//Make that catalog dir if it doesn't exist
	catalogpath := filepath.Join("downloads", catalog)
	if _, err := os.Stat(catalogpath); os.IsNotExist(err) {
		err := os.Mkdir(catalogpath, 0777)
		checkError(err)
	}

	sp = filepath.Join(catalogpath, sp)
	return sp

}

//check if the url is valid
func checkURL(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}

//download the image from the url
func downloadURL(url, filename string) {

	resp, err := http.Get(url)
	checkError(err)

	f, err := os.Create(filename)
	checkError(err)

	io.Copy(f, resp.Body)

	resp.Body.Close()
	f.Close()

}

type Capabilities struct {
	XMLName               xml.Name `xml:"Capabilities"`
	Text                  string   `xml:",chardata"`
	Xmlns                 string   `xml:"xmlns,attr"`
	Ows                   string   `xml:"ows,attr"`
	Xlink                 string   `xml:"xlink,attr"`
	Xsi                   string   `xml:"xsi,attr"`
	Gml                   string   `xml:"gml,attr"`
	SchemaLocation        string   `xml:"schemaLocation,attr"`
	Version               string   `xml:"version,attr"`
	ServiceIdentification struct {
		Text               string `xml:",chardata"`
		Title              string `xml:"Title"`
		ServiceType        string `xml:"ServiceType"`
		ServiceTypeVersion string `xml:"ServiceTypeVersion"`
	} `xml:"ServiceIdentification"`
	OperationsMetadata struct {
		Text      string `xml:",chardata"`
		Operation []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
			DCP  struct {
				Text string `xml:",chardata"`
				HTTP struct {
					Text string `xml:",chardata"`
					Get  struct {
						Text       string `xml:",chardata"`
						Href       string `xml:"href,attr"`
						Constraint struct {
							Text          string `xml:",chardata"`
							Name          string `xml:"name,attr"`
							AllowedValues struct {
								Text  string `xml:",chardata"`
								Value string `xml:"Value"`
							} `xml:"AllowedValues"`
						} `xml:"Constraint"`
					} `xml:"Get"`
				} `xml:"HTTP"`
			} `xml:"DCP"`
		} `xml:"Operation"`
	} `xml:"OperationsMetadata"`
	Contents struct {
		Text  string `xml:",chardata"`
		Layer struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"Title"`
			Identifier  string `xml:"Identifier"`
			BoundingBox struct {
				Text        string `xml:",chardata"`
				Crs         string `xml:"crs,attr"`
				LowerCorner string `xml:"LowerCorner"`
				UpperCorner string `xml:"UpperCorner"`
			} `xml:"BoundingBox"`
			WGS84BoundingBox struct {
				Text        string `xml:",chardata"`
				Crs         string `xml:"crs,attr"`
				LowerCorner string `xml:"LowerCorner"`
				UpperCorner string `xml:"UpperCorner"`
			} `xml:"WGS84BoundingBox"`
			Style struct {
				Text       string `xml:",chardata"`
				IsDefault  string `xml:"isDefault,attr"`
				Title      string `xml:"Title"`
				Identifier string `xml:"Identifier"`
			} `xml:"Style"`
			Format            string `xml:"Format"`
			TileMatrixSetLink struct {
				Text          string `xml:",chardata"`
				TileMatrixSet string `xml:"TileMatrixSet"`
			} `xml:"TileMatrixSetLink"`
			ResourceURL struct {
				Text         string `xml:",chardata"`
				Format       string `xml:"format,attr"`
				ResourceType string `xml:"resourceType,attr"`
				Template     string `xml:"template,attr"`
			} `xml:"ResourceURL"`
		} `xml:"Layer"`
		TileMatrixSet struct {
			Text         string `xml:",chardata"`
			Title        string `xml:"Title"`
			Abstract     string `xml:"Abstract"`
			Identifier   string `xml:"Identifier"`
			SupportedCRS string `xml:"SupportedCRS"`
			TileMatrix   []struct {
				Text             string `xml:",chardata"`
				Identifier       string `xml:"Identifier"`
				ScaleDenominator string `xml:"ScaleDenominator"`
				TopLeftCorner    string `xml:"TopLeftCorner"`
				TileWidth        string `xml:"TileWidth"`
				TileHeight       string `xml:"TileHeight"`
				MatrixWidth      string `xml:"MatrixWidth"`
				MatrixHeight     string `xml:"MatrixHeight"`
			} `xml:"TileMatrix"`
		} `xml:"TileMatrixSet"`
	} `xml:"Contents"`
	ServiceMetadataURL struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
	} `xml:"ServiceMetadataURL"`
}

// Loads a catalog from wmtsXML
func LoadCatalog(wmtsXML string) Capabilities {

	checkURL(wmtsXML)
	data, err := http.Get(wmtsXML)
	checkError(err)

	body, err := ioutil.ReadAll(data.Body)
	checkError(err)
	defer data.Body.Close()

	var capabilities Capabilities
	xml.Unmarshal(body, &capabilities)

	return capabilities
}

// Fetch an exact dataset pertaining to a specific LOD
func FetchExact(xmlURL string, LOD int, wg *sync.WaitGroup, uiPbar *widget.ProgressBar) bool {

	var misses []string
	start := time.Now()

	checkURL(xmlURL)
	data, err := http.Get(xmlURL)
	checkError(err)

	body, err := ioutil.ReadAll(data.Body)
	checkError(err)

	var capabilities Capabilities
	xml.Unmarshal(body, &capabilities)

	style := capabilities.Contents.Layer.Style.Identifier

	// PROBLEM HERE...

	matrixWidth := (capabilities.Contents.TileMatrixSet.TileMatrix[LOD].MatrixWidth)
	matrixHeight := (capabilities.Contents.TileMatrixSet.TileMatrix[LOD].MatrixHeight)

	tileMatrixSet := capabilities.Contents.TileMatrixSet.Identifier
	url := capabilities.Contents.Layer.ResourceURL.Template

	height, err := strconv.ParseFloat(matrixHeight, 64)
	checkError(err)

	width, err := strconv.ParseFloat(matrixWidth, 64)
	checkError(err)

	// ------ DEBUG -----
	//log.Println("Title:", capabilities.Contents.TileMatrixSet.Title)
	//log.Println("Catalog:", capabilities.Contents.Layer.Identifier)
	//log.Println("Format:", capabilities.Contents.Layer.Format)
	//log.Println("TileMatrixSetID:", capabilities.Contents.Layer.TileMatrixSetLink.TileMatrixSet)
	//log.Println("Style:", capabilities.Contents.Layer.Style.Identifier)
	//log.Println("WMTS matrixWidth :", matrixWidth)
	//log.Println("WMTS matrixHeight:", matrixHeight)
	//log.Println("TileMatrixSet:", capabilities.Contents.TileMatrixSet.Identifier)
	//log.Println("URL: ", url)

	var progBarLimit int64 = int64(width * height)
	bar := progressbar.Default(progBarLimit)

	log.Println("Total Tiles to fetch:", width*height)
	for row := 0; row < int(height); row++ {
		// Cli and UI progressbars both need ticking
		bar.Add(1)
		uiPbar.SetValue(float64(row) + 1.0)

		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			for col := 0; col < int(width); col++ {
				// Cli and UI progressbars both need ticking
				bar.Add(1)
				uiPbar.SetValue(float64(row) + 1.0)

				catalog := capabilities.Contents.Layer.Identifier
				url := buildURL(url, style, tileMatrixSet, &LOD, row, col)
				filename := makeFilenameFromURL(url, catalog)
				if !isAlreadyDownloaded(filename) {
					if checkURL(url) {
						downloadURL(url, filename)

					} else {
						log.Println("URL not valid:", url)
						misses = append(misses, url)

					}
				} else {
					log.Println("File already exists:", filename)
				}
			}

		}(row)
		wg.Wait() //NOTE: unsure whether to place this here, or one layer further down...
	}

	elapsed := time.Since(start)
	log.Println("Time to download:", elapsed.Minutes(), "mins")
	if len(misses) > 0 {
		return false
	} else {
		return true
	}

}

// keep files we missed for another time..
func writeMisses(misses []string) {
	f, err := os.Create("misses.txt")
	checkError(err)
	w := bufio.NewWriter(f)
	for _, miss := range misses {
		w.WriteString(miss + "\n")
	}
	w.Flush()
	f.Close()
}

// Download the image from the urls in misses.txt
func FetchMisses() []string {
	f, err := os.Open("misses.txt")
	checkError(err)

	var urls []string
	// append the urls to the urls slice
	scanner := bufio.NewScanner(f)
	f.Close()
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	return urls

}

// DEBUG for some time now we've been having too-many-open files errors
func countOpenFiles() int64 {
	out, err := exec.Command("/bin/zsh", "-c", fmt.Sprintf("lsof -p %v", os.Getpid())).Output()
	checkError(err)
	lines := strings.Split(string(out), "\n")
	return int64(len(lines) - 1)
}
