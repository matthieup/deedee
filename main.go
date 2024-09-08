package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type NetstatEntry struct {
	Source      string
	Destination string
}

type NetstatEntryDetail struct {
	SrcHost   string
	SrcPort   string
	DstHost   string
	DstPort   string
	Direction string
}

const Regexp = `\S+ *\d+ *\d+ *(\S+) *(\S+) *ESTABLISHED *`

var output []byte

func formatNetstats(allNetstat []NetstatEntry) *bytes.Buffer {
	w := bytes.NewBuffer(output)
	for _, v := range allNetstat {
		src_hostname_port := strings.Split(v.Source, ":")
		dst_hostname_port := strings.Split(v.Destination, ":")
		SrcPort := src_hostname_port[1]
		SrcHost := src_hostname_port[0]
		DstPort := dst_hostname_port[1]
		DstHost := dst_hostname_port[0]
		Direction := ""
		if i, _ := strconv.Atoi(SrcPort); i >= 20000 {
			Direction = "->"
			SrcPort = ""
		} else {
			Direction = "<-"
			DstPort = ""
		}

		templateD2, err := template.ParseFiles("d2template.j2")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		templateD2.Execute(w, &NetstatEntryDetail{SrcHost, SrcPort, DstHost, DstPort, Direction})
	}
	return w
}

func main() {
	allNetstat := []NetstatEntry{}

	var outputFileName = flag.String("o", "out.d2", "D2 Output filename")
	flag.Parse()

	cmd := exec.Command("/usr/bin/sudo", "netstat", "-alp", "-v4")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	all_lines := strings.Split(string(output), "\n")
	for _, v := range all_lines {
		if strings.Contains(v, "ESTABLISHED") &&
			!strings.Contains(v, "localhost") {
			comp := regexp.MustCompile(Regexp)
			re := comp.FindStringSubmatch(v)
			if err != nil {
				fmt.Println("Could not find")
			} else {
				newNetstat := NetstatEntry{
					Source:      re[1],
					Destination: re[2],
				}
				allNetstat = append(allNetstat, newNetstat)
			}
		}
	}
	outputB := formatNetstats(allNetstat)

	os.Remove(*outputFileName)
	outFile, err := os.Create(*outputFileName)
	if err != nil {
		fmt.Println("error writing to out.d2")
	}
	outFile.Write(outputB.Bytes())
}
