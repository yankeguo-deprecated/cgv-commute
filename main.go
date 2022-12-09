package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"github.com/guoyk93/gg"
	"github.com/guoyk93/gg/ggos"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	FileListTXT     = "list.txt"
	FileCombinedMP4 = "combined.mp4"
	FileFinalMP4    = "final.mp4"
	FileCoverRawPNG = "cover-raw.png"
	FileCoverJPG    = "cover.jpg"
	DirRes          = "res"
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
		optTo  bool
		optOff bool
	)

	flag.BoolVar(&optTo, "to", false, "to work")
	flag.BoolVar(&optOff, "off", false, "off work")
	flag.Parse()

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
	gg.Must0(os.RemoveAll(FileListTXT))
	gg.Must0(os.RemoveAll(FileCombinedMP4))
	gg.Must0(os.RemoveAll(FileFinalMP4))
	gg.Must0(os.RemoveAll(FileCoverRawPNG))
	gg.Must0(os.RemoveAll(FileCoverJPG))

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

	// build list
	{
		out := &bytes.Buffer{}
		for _, item := range gg.Must(os.ReadDir(".")) {
			if item.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(item.Name()), ".mp4") {
				continue
			}
			gg.Must(out.WriteString("file '" + item.Name() + "'\n"))
		}
		gg.Must0(os.WriteFile(FileListTXT, out.Bytes(), 0640))
		defer os.RemoveAll(FileListTXT)
	}

	// combine file
	{
		gg.Must0(
			execute(
				"ffmpeg",
				"-f",
				"concat",
				"-safe",
				"0",
				"-i",
				FileListTXT,
				"-c",
				"copy",
				FileCombinedMP4,
			),
		)
		defer os.RemoveAll(FileCombinedMP4)
	}

	// final
	{
		gg.Must0(
			execute(
				"ffmpeg",
				"-i",
				FileCombinedMP4,
				"-i",
				filepath.Join("res", "overlay.png"),
				"-filter_complex",
				"[0:v][1:v] overlay=0:0",
				"-c:v",
				"h264_videotoolbox",
				"-b:v",
				"15M",
				"-an",
				FileFinalMP4,
			),
		)
		os.RemoveAll(FileCombinedMP4)
	}

	// snapshot
	{
		gg.Must0(
			execute(
				"ffmpeg",
				"-ss",
				"00:00:05",
				"-i",
				FileFinalMP4,
				"-vframes",
				"1",
				"-q:v",
				"1",
				FileCoverRawPNG,
			),
		)
		defer os.RemoveAll(FileCoverRawPNG)
	}

	{
		gg.Must0(
			execute(
				"ffmpeg",
				"-i",
				FileCoverRawPNG,
				"-i",
				filepath.Join("res", "cover-overlay-"+mode+"-work.png"),
				"-filter_complex",
				"[0:v][1:v] overlay=0:0",
				FileCoverJPG,
			),
		)

		os.RemoveAll(FileCoverRawPNG)
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
