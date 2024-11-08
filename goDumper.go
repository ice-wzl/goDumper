package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func getMaps(port string) []string {
	// Lots of credit to https://github.com/BishopFox/sliver/blob/master/implant/sliver/procdump/dump_linux.go

	var targetMaps []string

	memFile := fmt.Sprintf("/proc/%s/maps", port)

	file, err := os.OpenFile(memFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("[!] Error opening file:", memFile)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")

		// The vvar region is for shared kernel variables
		// Other regions reserved by the kernel proved problematic, vdso and vsyscall
		region := parts[len(parts)-1]
		if region == "[vvar]" || region == "[vdso]" || region == "[vsyscall]" {
			continue
		}
		isFile := parts[3]

		// Then this is not a file
		if isFile == "00:00" {
			continue
		}

		// fmt.Println("Debug region:", region)
		maps := parts[0]
		targetMaps = append(targetMaps, maps)
	}

	return targetMaps

}

func getStartStop(memRange string) (int64, int64) {
	rangeMem := strings.Split(memRange, "-") // - dividing the start and the stop addresses
	memStart := rangeMem[0]
	memStop := rangeMem[1]
	memStartNum, _ := strconv.ParseInt(memStart, 16, 64)
	memEndNum, _ := strconv.ParseInt(memStop, 16, 64)

	return memStartNum, memEndNum
}

func doDump(memStart int64, memEnd int64, pid int) {

	memFileName := fmt.Sprintf("/proc/%v/mem", pid)
	pwd, err := os.Getwd()

	if err != nil {
		fmt.Println("[!] Failed to get pwd:", err)
		os.Exit(1)
	}

	outputName := fmt.Sprintf("%v/dump.%v", pwd, pid)

	outFile, err := os.OpenFile(outputName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[!] Failed to create dumpfile:", err)
		os.Exit(1)
	}
	defer outFile.Close()

	memFile, err := os.OpenFile(memFileName, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("[!] Failed to open memFile:", memFile)
		fmt.Println("\t--> Are you the root user...")
		os.Exit(1)
	}

	_, err = memFile.Seek(memStart, io.SeekStart)
	if err != nil {
		fmt.Println("[!] Failed to seek to file start:", err)
		os.Exit(1)
	}

	bufferSize := int64(4096)
	buffer := make([]byte, bufferSize)

	bytesRemaining := memEnd - memStart
	for bytesRemaining > 0 {
		bytesToRead := bufferSize
		if bytesRemaining < bufferSize {
			bytesToRead = bytesRemaining
		}

		// Read the memory into the buffer
		n, err := memFile.Read(buffer[:bytesToRead])
		if err != nil && err != io.EOF {

			fmt.Printf("[!] Skipping, failed to read memory: %x-%x: %v\n", memStart, memEnd, err)
			return
		}

		// Write the buffer content to the output file
		if _, err := outFile.Write(buffer[:n]); err != nil {
			fmt.Printf("[!] Failed to write to output file: %v\n", err)
			return
		}

		bytesRemaining -= int64(n)
		if err == io.EOF {
			break
		}
	}
}

func main() {

	pid := flag.String("p", "", "The pid of the process to memory dump")
	singleShotRange := flag.String("r", "", "[Optional] The single memory range to target -> 77535b8d5000-77535b8d7000")
	flag.Parse()
	fmt.Println("[+] goDumper started")

	if *pid == "" {
		fmt.Println("[!] PID must be specified in order to dump process memory")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("[+] Target PID:", *pid)

	// full memory dump of the target pid
	if *pid != "" && *singleShotRange == "" {
		// Getting maps
		targetMaps := getMaps(*pid)

		pidInt, err := strconv.Atoi(*pid)
		if err != nil {
			fmt.Println("[!] Error converting target pid to int:", err)
			os.Exit(1)
		}

		for _, line := range targetMaps {
			memStart, memEnd := getStartStop(line)
			doDump(memStart, memEnd, pidInt)

		}
		fmt.Println("[+] Successful memory dump for pid:", pidInt)
	} else if *pid != "" && *singleShotRange != "" {
		pidInt, err := strconv.Atoi(*pid)
		if err != nil {
			fmt.Println("[!] Error converting target pid to int:", err)
			os.Exit(1)
		}
		memStart, memEnd := getStartStop(*singleShotRange)
		doDump(memStart, memEnd, pidInt)
		fmt.Println("[+] Successful memory dump for pid:", pidInt)
	}

}
