package tools

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"log"

	"github.com/inancgumus/screen"
	"github.com/schollz/progressbar"
	"gocv.io/x/gocv"
)

// check the size of the image
func checkSize(path string) bool {
	img := gocv.IMRead(path, gocv.IMReadColor)
	if img.Size()[0] == 256 {
		return true
	}
	return false
}

func ConcatWithPython(tp string, lod int, format string) {
	lodString := strconv.Itoa(lod)
	cmd := exec.Command("python3", "scripts/stitcher.py", tp, lodString, format)
	log.Println(cmd)
	res, err := cmd.Output()
	if err != nil {
		log.Println("Call to python failed:", err)
		log.Fatal("resulting in:", res)
	}

}

func RunConcatenations(depth, dir string) {
	// Run the concatenation process on the directory, at the LOD/depth passed
	LOD, err := strconv.Atoi(depth)
	if err != nil {
		log.Fatal(err)
	}

	// somewhere to store intermediate results
	createTmp()

	catalog := strings.Split(dir, "/")[1]

	//check extension
	readDir, _ := ioutil.ReadDir(dir)
	ext := readDir[0].Name()[len(readDir[0].Name())-4:]
	//check img sizes
	for _, f := range readDir {
		if !checkSize(filepath.Join(dir, f.Name())) {
			log.Fatal("bad size:", f)
			panic("bad size on " + f.Name())
		}
	}

	// LOD Table for EquiRectangular, which not all datasets will be..
	// LOD|Height/Cols|Width/Rows
	//   2         4           8
	//   3         8           16
	//   4         16          32
	//   5         32          64
	//   6         64          128
	//   7         128         256
	// of all data explored thusfar, 256x256 is the only resolution that a tile is/can-be...
	var Width int
	var Height int
	switch LOD {
	case 2:
		Height = 4
		Width = 8
	case 3:
		Height = 8
		Width = 16
	case 4:
		Height = 16
		Width = 32
	case 5:
		Height = 32
		Width = 64
	case 6:
		Height = 64
		Width = 128
	case 7:
		Height = 128
		Width = 256
	}

	bar1 := progressbar.New(Height * Width)
	log.Println("Width:", Width, "Height:", Height)

	//Make horizontal slices
	for r := 0; r < Height; r++ {
		filename := filepath.Join(dir, strconv.Itoa(LOD)+"_"+strconv.Itoa(r)+"_"+strconv.Itoa(0)+ext)
		hfin := gocv.IMRead(filename, gocv.IMReadColor)
		if r == Height {
			break
		}
		for c := 1; c < Width; c++ {
			if c == Width {
				break
			}
			filename = filepath.Join(dir, strconv.Itoa(LOD)+"_"+strconv.Itoa(r)+"_"+strconv.Itoa(c)+ext)
			src2 := gocv.IMRead(filename, gocv.IMReadColor)
			mat := gocv.NewMat()
			gocv.Hconcat(hfin, src2, &mat)
			hfin = mat

			bar1.Add(1)
		}
		bar1.Add(1)
		outpath := filepath.Join("tmp", strconv.Itoa(r)+ext)
		gocv.IMWrite(outpath, hfin)

	}

	screen.MoveTopLeft()
	screen.Clear()
	log.Println("\n*********************************")

	hSlices, err := ioutil.ReadDir(filepath.Join("tmp"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Num Slices:", hSlices)

	// Vertitacally concatenate the horizontal slices, loading them from disk
	fin := gocv.IMRead(filepath.Join("tmp", "0"+ext), gocv.IMReadColor)

	bar2 := progressbar.New(len(hSlices))
	for rdx := 1; rdx < len(hSlices)-1; rdx++ {
		bar2.Add(1)
		filename := strconv.Itoa(rdx) + ext
		mat := gocv.NewMat()
		src2 := gocv.IMRead(filepath.Join("tmp", filename), gocv.IMReadColor)
		gocv.Vconcat(fin, src2, &mat)
		fin = mat
	}

	gocv.IMWrite("stitched_results/"+catalog+"_"+strconv.Itoa(LOD)+ext, fin)

	log.Println("Final size:", fin.Size()[0], fin.Size()[1])

	// DEBUG: Show the final img
	window := gocv.NewWindow("MARS")
	window.ResizeWindow(640, 480)
	// TODO: draw the dimms on the one we show for debug and its filepath...
	window.IMShow(fin)
	window.WaitKey(0)
	if window.WaitKey(1) >= 0 {
		window.Close()
		return
	}

	cleanTmp()
}

// remove tmp dir's contents
func cleanTmp() {
	os.RemoveAll(filepath.Join("tmp"))

}

// if the tmp/ dir does not exist, create it
func createTmp() {
	if _, err := os.Stat(filepath.Join("tmp")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join("tmp"), 0777)
	} else {
		log.Println("tmp/ exists")
	}

}

// a struct to hold the row/col of a tile
type tilePos struct {
	row int
	col int
}

// Return the row/col valus of the tile from its path
func getRowCol(imgPath string) tilePos {
	sp := strings.Split(imgPath, "_")
	// No reason to believe these will fail to parse if we're getting them from disk
	row, _ := strconv.Atoi(sp[1])
	col, _ := strconv.Atoi(strings.Split(sp[2], ".")[0])
	return tilePos{row, col}
}
