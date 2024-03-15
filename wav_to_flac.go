package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	var INPUT_FOLDER string
	var OUTPUT_FOLDER string

	var fileExtension string
	var isMono string

	arguments := os.Args[1:]

	INPUT_FOLDER = arguments[0]
	OUTPUT_FOLDER = arguments[1]

	fmt.Println("wav to flac started")

	err := filepath.Walk(INPUT_FOLDER,
		func(pathh string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			createDirectory(pathh, INPUT_FOLDER, OUTPUT_FOLDER)
			fileExtension = strings.ToLower(path.Ext(pathh))

			if fileExtension == ".wav" || fileExtension == ".aif" || fileExtension == ".aiff" || fileExtension == ".flac" || fileExtension == ".mp3" {

				isMono = sox_mono_checker(pathh)
				fmt.Println(pathh, isMono)
				if isMono == "mono" {
					ffmpegConverter(pathh, INPUT_FOLDER, OUTPUT_FOLDER, "1")
				} else if isMono == "stereo" {
					ffmpegConverter(pathh, INPUT_FOLDER, OUTPUT_FOLDER, "2")
				}
			}

			return nil
		})
	if err != nil {
		fmt.Println("error walk")

	}
}

func createDirectory(fullpath string, input_folder string, output_folder string) {
	outputPath := strings.Replace(fullpath, input_folder, output_folder, 1)

	fileInfo, err := os.Stat(fullpath)
	if err != nil {
		fmt.Println("fileInfo os.Stat", err)
	}

	if fileInfo.IsDir() {
		err := os.Mkdir(outputPath, 0755)
		if err == nil || os.IsExist(err) {
			fmt.Println("dir already exists")
		} else {
			fmt.Println(err)
		}
	}
}

func ffmpegConverter(fullpath string, input_folder string, output_folder string, channels string) {
	outputPath := strings.Replace(fullpath, input_folder, output_folder, 1)

	fileExtension := path.Ext(outputPath)
	outputPath = strings.Replace(outputPath, fileExtension, ".flac", 1)

	fmt.Println(outputPath)
	fmt.Println(channels)
	ffCmd := exec.Command("ffmpeg", "-i", fullpath, "-ac", channels, "-af", "aformat=s16:44100", outputPath)
	ffCmd.Stdout = os.Stdout
	ffCmd.Stderr = os.Stderr
	err := ffCmd.Run()
	if err != nil {
		fmt.Println("ffCmd error", err)
	}

}

func sox_mono_checker(inputFile string) string {
	var tmpMaxAmpString string
	var soxMaximumAmplitude float64
	soxMaximumAmplitude = 0
	soxCmd := exec.Command("sox", inputFile, "-n", "oops", "stat")
	var stdout, stderr bytes.Buffer
	soxCmd.Stdout = &stdout
	soxCmd.Stderr = &stderr
	err := soxCmd.Run()
	if err != nil {

		if strings.Contains(stderr.String(), "too few input channels") {
			//fmt.Println("sample is mono")
			return "mono"
		} else {
			//fmt.Println("error")
			return "broken"

		}
	}

	soxOutput := stderr.String()
	//fmt.Println(soxOutput)
	soxTmp := bufio.NewScanner(strings.NewReader(soxOutput))
	for soxTmp.Scan() {
		if strings.Contains(soxTmp.Text(), "Maximum amplitude") {
			tmpMaxAmpString = soxTmp.Text()
		}
	}

	tmpMaxAmpString = strings.Replace(tmpMaxAmpString, "Maximum amplitude:", "", 1)
	tmpMaxAmpString = strings.Replace(tmpMaxAmpString, " ", "", -1)

	soxMaximumAmplitude, err = strconv.ParseFloat(tmpMaxAmpString, 32)
	if err != nil {
		fmt.Println("some error string to float")
		return "broken"
	}

	if soxMaximumAmplitude > 0.00001 {

		return "stereo"
	} else {

		return "mono"
	}

}
