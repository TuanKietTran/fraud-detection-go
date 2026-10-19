package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	profile := "coverage.out"
	if len(os.Args) > 1 {
		profile = os.Args[1]
	}
	file, err := os.Open(profile)
	if err != nil {
		fail(err.Error())
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var total, covered, domainTotal, domainCovered int64
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			fail("invalid coverage profile line: " + line)
		}
		var statements, count int64
		if _, err := fmt.Sscan(fields[1], &statements); err != nil || statements <= 0 {
			fail("invalid statement count: " + line)
		}
		if _, err := fmt.Sscan(fields[2], &count); err != nil || count < 0 {
			fail("invalid execution count: " + line)
		}
		if !strings.Contains(fields[0], "/cmd/") {
			total += statements
			if count > 0 {
				covered += statements
			}
		}
		if strings.Contains(fields[0], "/internal/domain/") {
			domainTotal += statements
			if count > 0 {
				domainCovered += statements
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fail(err.Error())
	}
	if total == 0 || domainTotal == 0 {
		fail("coverage profile must contain non-cmd and domain statements")
	}
	fmt.Printf("weighted non-cmd: %d/%d = %.2f%%\n", covered, total, 100*float64(covered)/float64(total))
	fmt.Printf("domain: %d/%d = %.2f%%\n", domainCovered, domainTotal, 100*float64(domainCovered)/float64(domainTotal))
	if 100*covered < 80*total || 100*domainCovered < 90*domainTotal {
		os.Exit(1)
	}
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
