package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/donuts-are-good/colors"
)

type Specs struct {
	Userhost   string
	OS         string
	Kernel     string
	Uptime     string
	Shell      string
	CPU        string
	RAM        string
	GPU        string
	SystemArch string
	DiskUsage  string
}

var (
	noColors bool
	help     bool
)

func init() {
	flag.BoolVar(&noColors, "nocolors", false, "Disable colored output")
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.Parse()
}

func main() {

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
			disableColors()
	}
	
	if noColors {
		disableColors()
	}
	if help {
		showHelp()
		return
	}
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
	fmt.Println(colors.BrightMagenta + newInfo.Userhost + colors.Nc)
	fmt.Println("------------")
	fmt.Println(colors.Cyan+"OS:     "+colors.Nc, info.OS)
	fmt.Println(colors.Cyan+"Kernel: "+colors.Nc, info.Kernel)
	fmt.Println(colors.Cyan+"Uptime: "+colors.Nc, info.Uptime)
	fmt.Println(colors.Cyan+"Shell:  "+colors.Nc, info.Shell)
	fmt.Println(colors.Cyan+"CPU:    "+colors.Nc, info.CPU)
	fmt.Println(colors.Cyan+"RAM:    "+colors.Nc, info.RAM)
	fmt.Println(colors.Cyan+"GPU:    "+colors.Nc, info.GPU)
	fmt.Println(colors.Cyan+"Arch:   "+colors.Nc, info.SystemArch)
	fmt.Println(colors.Cyan+"Disk:   "+colors.Nc, info.DiskUsage)
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
	info.GPU, _ = getGPUInfo()
	info.SystemArch, _ = getSystemArch()
	info.DiskUsage, _ = getDiskUsage()
	infoChan <- *info
}

func showHelp() {
	fmt.Println(`
Usage: gofetch [OPTIONS]

gofetch is a tool to display system information, such as OS, kernel, uptime, shell, CPU, RAM, GPU, architecture, and disk usage.

OPTIONS:
  -nocolors     Disable colored output

Example:
  gofetch -nocolors`)
}

func disableColors() {
	colors.BrightMagenta = ""
	colors.Cyan = ""
	colors.Nc = ""
}

func getGPUInfo() (string, error) {
	var output []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		output, err = exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on Windows: %v", err)
		}
	case "darwin":
		output, err = exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on macOS: %v", err)
		}
	case "linux":
		output, err = exec.Command("lspci", "-vnn").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on Linux: %v", err)
		}
	default:
		return "", fmt.Errorf("error: GPU information retrieval not implemented for %s", runtime.GOOS)
	}

	outputStr := strings.TrimSpace(string(output))

	if runtime.GOOS == "windows" {
		lines := strings.Split(outputStr, "\r\n")[1:]
		gpuName := strings.TrimSpace(strings.Join(lines, " "))
		return gpuName, nil
	}

	if runtime.GOOS == "darwin" {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Chipset Model:") {
				fields := strings.Split(line, ":")
				if len(fields) >= 2 {
					gpuName := strings.TrimSpace(fields[1])
					return gpuName, nil
				}
			}
		}
		return "", fmt.Errorf("error parsing GPU information on macOS")
	}

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA compatible controller") {
			fields := strings.Fields(line)
			if len(fields) > 2 {
				gpuName := strings.Join(fields[2:], " ")
				return gpuName, nil
			}
		}
	}
	return "", fmt.Errorf("error parsing GPU information on Linux")
}

func getSystemArch() (string, error) {
	arch := runtime.GOARCH
	if arch == "" {
		return "", fmt.Errorf("unable to determine system architecture")
	}
	return arch, nil
}

func getDiskUsage() (string, error) {
	var output []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		output, err = exec.Command("wmic", "logicaldisk", "where", "drivetype=3", "get", "size,freespace").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving disk usage on Windows: %v", err)
		}
	case "darwin", "linux":
		output, err = exec.Command("df", "-h", "-t").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving disk usage on %s: %v", runtime.GOOS, err)
		}
	default:
		return "", fmt.Errorf("error: Disk usage retrieval not implemented for %s", runtime.GOOS)
	}

	outputStr := strings.TrimSpace(string(output))

	if runtime.GOOS == "windows" {
		lines := strings.Split(outputStr, "\r\n")[1:]
		var totalSize, totalFree uint64
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				size, _ := strconv.ParseUint(fields[0], 10, 64)
				free, _ := strconv.ParseUint(fields[1], 10, 64)
				totalSize += size
				totalFree += free
			}
		}
		return fmt.Sprintf(colors.Cyan+"Total: "+colors.Nc+"%d GB, Free: %d GB", totalSize/(1024*1024*1024), totalFree/(1024*1024*1024)), nil
	}

	if runtime.GOOS == "darwin" {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 4 && strings.Contains(fields[0], "/dev/") && fields[8] == "/" {
				size, err := strconv.ParseFloat(strings.TrimRight(fields[1], "BGMi"), 64)
				if err != nil {
					return "", fmt.Errorf("error parsing size value: %v", err)
				}
				avail, err := strconv.ParseFloat(strings.TrimRight(fields[3], "BGMi"), 64)
				if err != nil {
					return "", fmt.Errorf("error parsing avail value: %v", err)
				}
				used := size - avail
				return fmt.Sprintf("Total: "+colors.Nc+"%.1fG\n"+colors.Cyan+"Disk:    "+colors.Nc+"Free:  %.1fG\n"+colors.Cyan+"Disk:    "+colors.Nc+"Used:  %.1fG", size, avail, used), nil
			}
		}
		return "", fmt.Errorf("error parsing disk usage on %s", runtime.GOOS)
	}

	lines := strings.Split(outputStr, "\n")
	totalLine := lines[len(lines)-2]
	fields := strings.Fields(totalLine)
	if len(fields) < 5 {
		return "", fmt.Errorf("error parsing disk usage on %s", runtime.GOOS)
	}
	return fmt.Sprintf(colors.Cyan+"Total: "+colors.Nc+"%s, Free: %s, Used: %s", fields[1], fields[3], fields[2]), nil
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
