package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type Specs struct {
	Userhost string
	OS       string
	Kernel   string
	Uptime   string
	Shell    string
	CPU      string
	RAM      string
}

func main() {
	info := &Specs{}
	infoChan := make(chan Specs, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go getSpecs(info, infoChan, &wg)
	wg.Wait()
	newInfo := <-infoChan
	info.Userhost = newInfo.Userhost
	info.OS = newInfo.OS
	info.Kernel = newInfo.Kernel
	info.Uptime = newInfo.Uptime
	info.Shell = newInfo.Shell
	info.CPU = newInfo.CPU
	info.RAM = newInfo.RAM
	fmt.Println(info.Userhost)
	fmt.Println("------------")
	fmt.Println("OS:     ", info.OS)
	fmt.Println("Kernel: ", info.Kernel)
	fmt.Println("Uptime: ", info.Uptime)
	fmt.Println("Shell:  ", info.Shell)
	fmt.Println("CPU:    ", info.CPU)
	fmt.Println("RAM:    ", info.RAM)
}

func getSpecs(info *Specs, infoChan chan Specs, wg *sync.WaitGroup) {
	defer wg.Done()
	info.Userhost = getUserHostname()
	info.OS = getOSName()
	info.Kernel = getKernelVersion()
	info.Uptime = getUptime()
	info.Shell = getShell()
	info.CPU = getCPUName()
	info.RAM = getMemStats()
	infoChan <- *info
}

func getUserHostname() string {
	hostname, _ := os.Hostname()
	return os.Getenv("USER") + "@" + hostname
}

func getOSName() string {
	return runtime.GOOS
}

func getKernelVersion() string {
	var kernelVersion string
	var err error
	switch runtime.GOOS {
	case "windows":
		output, err := exec.Command("ver").Output()
		if err != nil {
			fmt.Printf("Error retrieving kernel version on Windows: %v", err)
			return ""
		}
		kernelVersion = string(output)
	case "linux", "darwin":
		output, err := exec.Command("uname", "-r").Output()
		if err != nil {
			fmt.Printf("Error retrieving kernel version on %s: %v", runtime.GOOS, err)
			return ""
		}
		kernelVersion = string(output)
	case "freebsd", "openbsd", "netbsd":
		kernelVersion, err = syscall.Sysctl("kern.version")
		if err != nil {
			fmt.Printf("Error retrieving kernel version on BSD: %v", err)
			return ""
		}
	default:
		fmt.Printf("Error: Kernel version retrieval not implemented for %s", runtime.GOOS)
		return ""
	}

	return strings.TrimSpace(kernelVersion)
}

func getUptime() string {
	var uptime string
	switch runtime.GOOS {
	case "windows":
		output, err := exec.Command("net", "stats", "srv").Output()
		if err != nil {
			fmt.Printf("Error retrieving uptime on Windows: %v", err)
			return ""
		}
		outputStr := string(output)
		uptimeStart := strings.Index(outputStr, "Statistics since ") + 19
		uptimeEnd := strings.Index(outputStr[uptimeStart:], "\r\n")
		uptime = outputStr[uptimeStart : uptimeStart+uptimeEnd]
	case "linux":
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("Error retrieving uptime on Linux: %v", err)
			return ""
		}
		outputStr := string(output)
		uptimeStart := strings.Index(outputStr, "up ") + 3
		uptimeEnd := strings.Index(outputStr[uptimeStart:], ",")
		uptime = outputStr[uptimeStart : uptimeStart+uptimeEnd]
	case "darwin":
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("Error retrieving uptime on Darwin: %v", err)
			return ""
		}
		outputStr := string(output)
		uptimeStart := strings.Index(outputStr, "up ") + 3
		uptimeEnd := strings.Index(outputStr[uptimeStart:], ",")
		uptime = outputStr[uptimeStart : uptimeStart+uptimeEnd]
	case "freebsd", "openbsd", "netbsd":
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("Error retrieving uptime on BSD: %v", err)
			return ""
		}
		outputStr := string(output)
		uptimeStart := strings.Index(outputStr, "up ") + 3
		uptimeEnd := strings.Index(outputStr[uptimeStart:], ",")
		uptime = outputStr[uptimeStart : uptimeStart+uptimeEnd]
	default:
		fmt.Printf("Error: Uptime retrieval not implemented for %s", runtime.GOOS)
		return ""
	}

	return uptime
}

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "Unknown"
	}
	elements := strings.Split(shell, "/")
	shellName := elements[len(elements)-1]
	return shellName
}

func getCPUName() string {
	return runtime.GOARCH
}

func getMemStats() string {
	switch runtime.GOOS {
	case "darwin":
		output, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err != nil {
			fmt.Printf("Error retrieving memory info on Darwin: %v\n", err)
			return ""
		}
		outputStr := strings.TrimSpace(string(output))
		memSize, err := strconv.ParseUint(outputStr, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing memory info on Darwin: %v\n", err)
			return ""
		}
		return strconv.FormatUint(memSize/(1024*1024), 10) + "MB"
	case "freebsd", "openbsd", "netbsd":
		output, err := exec.Command("sysctl", "-n", "hw.physmem").Output()
		if err != nil {
			fmt.Printf("Error retrieving memory info on BSD: %v\n", err)
			return ""
		}
		outputStr := strings.TrimSpace(string(output))
		memSize, err := strconv.ParseUint(outputStr, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing memory info on BSD: %v\n", err)
			return ""
		}
		return strconv.FormatUint(memSize/(1024*1024), 10) + "MB"
	case "linux":
		output, err := exec.Command("free", "-m").Output()
		if err != nil {
			fmt.Printf("Error retrieving memory info on Linux: %v\n", err)
			return ""
		}
		outputStr := string(output)
		memIndex := strings.Index(outputStr, "Mem:")
		if memIndex == -1 {
			fmt.Println("Error parsing memory info on Linux")
			return ""
		}
		memLines := strings.Split(outputStr[memIndex:], "\n")[0]
		memFields := strings.Fields(memLines)
		if len(memFields) < 2 {
			fmt.Println("Error parsing memory info on Linux")
			return ""
		}
		totalRAM, err := strconv.ParseUint(memFields[1], 10, 64)
		if err != nil {
			fmt.Printf("Error parsing memory info on Linux: %v\n", err)
			return ""
		}
		return strconv.FormatUint(totalRAM, 10) + "MB"
	case "windows":
		output, err := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize").Output()
		if err != nil {
			fmt.Printf("Error retrieving memory info on Windows: %v\n", err)
			return "Unknown"
		}
		outputStr := strings.TrimSpace(string(output))
		memorySize, err := strconv.ParseUint(outputStr, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing memory size on Windows: %v\n", err)
			return "Unknown"
		}
		return strconv.FormatUint(memorySize/1024, 10) + "MB"
	default:
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		totalRAM := mem.TotalAlloc / (1024 * 1024)
		return strconv.FormatUint(totalRAM, 10) + "MB"
	}
}
