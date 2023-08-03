package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"
)

func main() {
	log.SetFlags(0)
	// Replace "logfile.log" with the path to your log file
	fileContent, err := ioutil.ReadFile("logfile.log")
	if err != nil {
		panic(err)
	}

	var newContent []byte
	for i := 0; i < len(fileContent); i++ {
		if fmt.Sprintf("%x", fileContent[i]) != "0" {
			newContent = append(newContent, fileContent[i])
		}
	}
	ioutil.WriteFile("logfile.log", newContent, 0)

	logLines := strings.Split(string(newContent), "\n")
	logCountMap := make(map[string]int)

	log.Printf("logLines leng : %d", len(logLines))

	for _, line := range logLines {
		if line != "" {
			logLineWithoutTime := extractLogLineWithoutTime(line)
			logCountMap[logLineWithoutTime]++
		}
	}

	// var mostRepeatedLogLine string
	// var maxCount int

	// for logLine, count := range logCountMap {
	// 	if count > maxCount {
	// 		mostRepeatedLogLine = logLine
	// 		maxCount = count
	// 	}
	// }

	// fmt.Println("Most repeated log line (ignoring time prefix):", mostRepeatedLogLine)
	// fmt.Println("Repeated", maxCount, "times.")
	// Create a slice to store log lines and counts
	type logLineCount struct {
		line  string
		count int
	}

	var logCountsSlice []logLineCount
	for line, count := range logCountMap {
		logCountsSlice = append(logCountsSlice, logLineCount{line, count})
	}

	// Sort the log lines by count (descending order)
	sort.SliceStable(logCountsSlice, func(i, j int) bool {
		return logCountsSlice[i].count > logCountsSlice[j].count
	})

	// Print the top 5 most repeated log lines
	fmt.Println("Top 5 most repeated log lines:")
	for i, logLine := range logCountsSlice {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s (repeated %d times)\n", i+1, logLine.line, logLine.count)
	}
}

func extractLogLineWithoutTime(logLine string) string {
	// Assuming the time format is fixed and always in the format "MM-dd HH:mm:ss.SSS"
	// Define the regular expression pattern to match the timestamp part
	pattern := "^\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3} "
	// pattern := "\\d{2}"

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Replace the timestamp part with an empty string
	cleanedLogLine := regex.ReplaceAllString(logLine, "")

	return cleanedLogLine
}
