package tools

import (
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"log"
)

func ReadApiEndpoints(filepath string) ([]string, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println(err)
	}
	xml := strings.Split(string(file), "\n")

	return xml[:len(xml)-1], nil

}
func ConcatWithPython(tp string, lod int) {
	lodString := strconv.Itoa(lod)
	cmd := exec.Command("python", "scripts/stitcher.py", tp, lodString)
	log.Println(cmd)
	res, err := cmd.Output()
	if err != nil {
		log.Println("Call to python failed:", err)
		log.Fatal("resulting in:", res)
	}

}
