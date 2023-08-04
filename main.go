package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"time"
)

type LogLineCount struct {
	line  string
	count int
}

func main() {
	log.SetFlags(0)
	// Replace "logfile.log" with the path to your log file
	logChannel := make(chan []string)

	go func() {
		log.Println("Start listener")
		filePath := "/mnt/f/MinGroup/MinSoftware/freelancer/Fiverr/AndroidManagerPasswordApp/logfile.log"

		// Track the last modification time of the file
		var lastModTime time.Time

		for {
			// Get the file information to check the modification time
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				log.Println("Error:", err)
				time.Sleep(time.Second) // Wait for 1 second before checking again
				continue
			}

			// Check if the file was modified since the last check
			if !fileInfo.ModTime().Equal(lastModTime) {
				// File was modified, read new lines

				// Open the file for reading
				file, err := os.Open(filePath)
				if err != nil {
					log.Println("Error:", err)
					time.Sleep(time.Second) // Wait for 1 second before checking again
					continue
				}

				// Move the file pointer to the last known position (to avoid reading the entire file)
				_, err = file.Seek(0, io.SeekEnd)
				if err != nil {
					log.Println("Error:", err)
					time.Sleep(time.Second) // Wait for 1 second before checking again
					continue
				}

				// Create a new scanner to read from the file
				scanner := bufio.NewScanner(file)

				// Read and process the newly added lines
				var logLines []string
				for scanner.Scan() {
					logLine := scanner.Text()
					logLines = append(logLines, string(cleanLine([]byte(logLine))))
				}
				logChannel <- logLines

				// Save the new modification time
				lastModTime = fileInfo.ModTime()
			}

			// Wait for a certain duration before checking again (e.g., 1 second)
			time.Sleep(time.Second)
		}
	}()

	log.Println("Start Consumer")

	logCounts := []LogLineCount{}
	var topLogLines []LogLineCount
	// Consumer
	for {

		newLines := <-logChannel
		for _, newLine := range newLines {
			if newLine != "" {
				base64Encoded := base64.StdEncoding.EncodeToString([]byte(newLine))
				if len(logCounts) == 0 {
					logCounts = append(logCounts, LogLineCount{
						line:  base64Encoded,
						count: 1,
					})
				} else {
					for i, logL := range logCounts {
						if logL.line == base64Encoded {
							logCounts[i].count++
							break
						}
						if i == len(logCounts)-1 {
							logCounts = append(logCounts, LogLineCount{
								line:  base64Encoded,
								count: 1,
							})
						}
					}
				}
			}
		}

		// Clear the current topLogLines
		topLogLines = topLogLines[:0]

		// // Populate topLogLines with the top 10 most repeated log lines
		for _, logLine := range logCounts {
			topLogLines = append(topLogLines, LogLineCount{logLine.line, logLine.count})
		}

		sort.SliceStable(topLogLines, func(i, j int) bool {
			return topLogLines[i].count > topLogLines[j].count
		})

		// // Print the top 10 most repeated log lines
		clearConsole()
		// log.Println("Top 10 most repeated log lines:")
		// for i, logLine := range topLogLines {
		// 	if i >= 10 {
		// 		break
		// 	}
		// 	log.Printf("(repeated %d times)\n", logLine.count)
		// }
		log.Printf("%d", len(logCounts))
		if len(topLogLines) > 0 {
			log.Printf("Top count %d, lowest count %d", topLogLines[0].count, topLogLines[len(topLogLines)-1].count)
			for _, logL := range logCounts {
				decoded, err := base64.StdEncoding.DecodeString(logL.line)
				if err == nil {
					log.Println(string(decoded))
				}
			}
		}
	}
}

func clearConsole() {
	cmd := exec.Command("clear") // Use "cls" instead of "clear" on Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func cleanLine(content []byte) []byte {
	var newContent []byte
	for i := 0; i < len(content); i++ {
		if fmt.Sprintf("%x", content[i]) != "0" {
			newContent = append(newContent, content[i])
		}
	}
	return newContent
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
