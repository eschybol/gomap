# Gomap
Ein NMAP Klon geschrieben in Golang

# Quickstart
```golang
go build .
./gomap -t <target_ip> -p <einzelne ports|portrange[80-445]> or -p- <alle ports>
```
# 2DO
	- Parallelization
		- routinen x
		- worker pool (load balancing) x
	- File Output
	- DNS Namensauflösung bei Hostnamen in der Eingabe
	- Annahme von Targetlisten
	- verschiedene Scanarten:
		- UDP Scan
		- Service Scan
		- Script Scan  
