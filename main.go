package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/guoyk93/gg"
	"github.com/guoyk93/gg/ggos"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	DirRes = "res"
)

var (
	//go:embed res
	RES embed.FS
)

func main() {
	var err error
	defer ggos.Exit(&err)
	defer gg.Guard(&err)

	var (
		optTo   bool
		optOff  bool
		optName string
	)

	flag.BoolVar(&optTo, "to", false, "to work")
	flag.BoolVar(&optOff, "off", false, "off work")
	flag.StringVar(&optName, "name", "", "name")
	flag.Parse()

	if optName == "" {
		err = errors.New("missing -name")
		return
	}

	fileFinal := optName + ".mp4"
	fileCover := optName + ".cover.jpg"

	var (
		mode string
	)

	if optTo {
		mode = "to"
	} else if optOff {
		mode = "off"
	} else {
		err = errors.New("one of 'on' and 'off' is required")
		return
	}

	_ = mode

	// clean files
	gg.Must0(os.RemoveAll(DirRes))
	defer gg.Must0(os.RemoveAll(DirRes))
	gg.Must0(os.RemoveAll(fileFinal))
	gg.Must0(os.RemoveAll(fileCover))

	// extract res
	{
		gg.Must0(os.MkdirAll(DirRes, 0750))
		for _, item := range gg.Must(RES.ReadDir(DirRes)) {
			if item.IsDir() {
				continue
			}
			buf := gg.Must(RES.ReadFile(path.Join(DirRes, item.Name())))
			gg.Must0(os.WriteFile(filepath.Join(DirRes, item.Name()), buf, 0640))
		}
	}

	// names
	var names []string
	{
		for _, item := range gg.Must(os.ReadDir(".")) {
			if item.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(item.Name()), ".mp4") {
				continue
			}
			names = append(names, item.Name())
		}
		sort.Strings(names)
	}

	if len(names) == 0 {
		err = errors.New("missing files")
		return
	}

	// create video
	{
		argv := []string{"ffmpeg"}
		for _, item := range names {
			argv = append(argv, "-i", item)
		}
		argv = append(argv, "-i", filepath.Join("res", "overlay.png"))
		argv = append(argv, "-i", filepath.Join("res", "title-overlay-"+mode+"-work.png"))

		idOverlay := len(names)
		idTitle := idOverlay + 1

		fcBuf := &bytes.Buffer{}

		fcBuf.WriteString(fmt.Sprintf("[0:v] [%d:v] overlay=enable='between(t,3,10)' [new0]; ", idTitle))

		fcBuf.WriteString("[new0] ")
		for i := range names {
			if i == 0 {
				continue
			}
			fcBuf.WriteString(fmt.Sprintf("[%d:v] ", i))
		}
		fcBuf.WriteString(fmt.Sprintf("concat=n=%d:v=1:a=0 [stage1]; ", len(names)))
		fcBuf.WriteString(fmt.Sprintf("[stage1] [%d:v] overlay [stage2]", idOverlay))

		argv = append(argv, "-filter_complex", fcBuf.String())
		argv = append(argv, "-map", "[stage2]")
		argv = append(argv, "-c:v", "h264_videotoolbox", "-b:v", "15M", fileFinal)

		gg.Must0(execute(argv...))
	}

	// snapshot
	{
		gg.Must0(
			execute(
				"ffmpeg",
				"-ss",
				"00:00:05",
				"-i",
				fileFinal,
				"-vframes",
				"1",
				"-q:v",
				"1",
				fileCover,
			),
		)
	}

}

func execute(argv ...string) (err error) {
	if len(argv) == 0 {
		err = errors.New("missing commands")
		return
	}
	gg.Log("execute: " + strings.Join(argv, " "))
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return
	}
	return
}
