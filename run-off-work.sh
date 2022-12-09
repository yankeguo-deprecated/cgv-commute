#!/bin/bash

set -eu

cd $(dirname $0)

rm -f videos.list.txt

ls -1 | grep MP4 | xargs -I {} bash -c "echo \"file '{}'\" >> videos.list.txt"

rm -f combined.mp4

ffmpeg -f concat -safe 0 -i videos.list.txt -c copy combined.mp4

rm -f final.mp4

ffmpeg -i combined.mp4 -i overlay.png -filter_complex '[0:v][1:v] overlay=0:0' -c:v h264_videotoolbox -b:v 15M -an final.mp4

rm -f combined.mp4

rm -f cover-raw.png

ffmpeg -ss 00:00:05 -i final.mp4 -vframes 1 -q:v 1 cover-raw.png

rm -f cover.png

ffmpeg -i cover-raw.png -i cover-overlay-off-work.png -filter_complex '[0:v][1:v] overlay=0:0' cover.png

rm -f cover-raw.png
