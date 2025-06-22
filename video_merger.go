package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var videoExtensions = []string{".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv"}

func isVideoFile(filename string) bool {
	if strings.Count(filename, ".") != 1 {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, v := range videoExtensions {
		if ext == v {
			return true
		}
	}
	return false
}

// 自然排序实现
type naturalStrings []string

func (ns naturalStrings) Len() int           { return len(ns) }
func (ns naturalStrings) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns naturalStrings) Less(i, j int) bool { return naturalLess(ns[i], ns[j]) }

var numRegexp = regexp.MustCompile(`(\d+)`)

func naturalLess(a, b string) bool {
	aa := numRegexp.FindAllStringIndex(a, -1)
	bb := numRegexp.FindAllStringIndex(b, -1)
	ai, bi := 0, 0
	for ai < len(aa) && bi < len(bb) {
		// 比较前缀
		if a[:aa[ai][0]] != b[:bb[bi][0]] {
			return a[:aa[ai][0]] < b[:bb[bi][0]]
		}
		// 比较数字
		numA, _ := strconv.Atoi(a[aa[ai][0]:aa[ai][1]])
		numB, _ := strconv.Atoi(b[bb[bi][0]:bb[bi][1]])
		if numA != numB {
			return numA < numB
		}
		// 截断前缀和数字，继续比较
		a = a[aa[ai][1]:]
		b = b[bb[bi][1]:]
		aa = numRegexp.FindAllStringIndex(a, -1)
		bb = numRegexp.FindAllStringIndex(b, -1)
		ai, bi = 0, 0
	}
	return a < b
}

func findVideoFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var videoFiles []string
	for _, file := range files {
		if !file.IsDir() && isVideoFile(file.Name()) {
			videoFiles = append(videoFiles, filepath.Join(dir, file.Name()))
		}
	}
	sort.Sort(naturalStrings(videoFiles))
	return videoFiles, nil
}

func createFileList(videoFiles []string, listPath string) error {
	f, err := os.Create(listPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, file := range videoFiles {
		fmt.Printf("Merging: %s\n", file)
		_, err := w.WriteString(fmt.Sprintf("file '%s'\n", file))
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func getFFmpegPath() string {
	// 优先查找当前目录下的 ffmpeg.exe（Windows）或 ffmpeg（其他平台）
	ffmpegName := "ffmpeg"
	if os.PathSeparator == '\\' {
		ffmpegName = "ffmpeg.exe"
	}
	cwd, err := os.Getwd()
	if err == nil {
		ffmpegPath := filepath.Join(cwd, ffmpegName)
		if _, err := os.Stat(ffmpegPath); err == nil {
			return ffmpegPath
		}
	}
	// 否则用系统 PATH
	return ffmpegName
}

func mergeVideos(videoFiles []string, output string) error {
	if len(videoFiles) == 0 {
		return fmt.Errorf("no video files to merge")
	}
	listPath := "video_list.txt"
	err := createFileList(videoFiles, listPath)
	if err != nil {
		return err
	}
	defer os.Remove(listPath)

	cmd := exec.Command(getFFmpegPath(), "-f", "concat", "-safe", "0", "-i", listPath, "-c:v", "libx264", "-c:a", "aac", "-strict", "-2", output)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// 实时输出 ffmpeg 的进度
	go io.Copy(os.Stdout, stdout)
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}

func main() {
	dir := flag.String("dir", "", "Directory containing video files to merge.")
	output := flag.String("output", "", "Filename for the merged output video.")
	flag.Parse()

	if *dir == "" || *output == "" {
		fmt.Println("Usage: video_merger_go --dir <video_dir> --output <output_file>")
		os.Exit(1)
	}

	videoFiles, err := findVideoFiles(*dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}
	if len(videoFiles) == 0 {
		fmt.Printf("No video files found in directory '%s'.\n", *dir)
		os.Exit(1)
	}

	fmt.Printf("Found %d video files to merge:\n", len(videoFiles))
	for _, vf := range videoFiles {
		fmt.Println("  -", vf)
	}

	err = mergeVideos(videoFiles, *output)
	if err != nil {
		fmt.Printf("Failed to merge videos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nSuccessfully merged videos into '%s'\n", *output)
}
