# gofetch
just a simple fetch thing for your terminal

## looks like this

```bash
╭─dh@lisa ~/Projects/gofetch ‹master●› 
╰─$ go build && ./gofetch           
dh@lisa.local
------------
OS:      darwin
Kernel:  22.1.0
Uptime:  7 days
Shell:   zsh
CPU:     arm64
RAM:     8192MB
GPU:     Apple M2
Arch:    arm64
Disk:    Total: 373.0G
Disk:    Free:  142.0G
Disk:    Used:  231.0G
```

## built like this
```bash
cd gofetch
go build
./gofetch
```

## supports --nocolors
```bash
./gofetch --nocolors
```