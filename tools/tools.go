package tools

import (
	"os/exec"
	"strconv"

	"log"
)

func ConcatWithPython(tp string, lod int, format string) {
	lodString := strconv.Itoa(lod)
	cmd := exec.Command("python", "scripts/stitcher.py", tp, lodString, format)
	log.Println(cmd)
	res, err := cmd.Output()
	if err != nil {
		log.Println("Call to python failed:", err)
		log.Fatal("resulting in:", res)
	}

}
