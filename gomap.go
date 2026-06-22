package main
import (
	"fmt"
	"flag"
	"net"
	"strings"
	"time"
	"os"
	"strconv"
	"sync"
)

/*
#2DO:
	- Parralelization
		- routinen x
		- worker pool (load balancing) x
	- Default port range: 1000 setzen
	- File Output
	- DNS Namensauflösung bei Hostnamen in der Eingabe
	- Annahme von Targetlisten
	- verschiedene Scanarten:
		- UDP Scan
		- Service Scan
		- Script Scan  
*/

var portList []int
var target, ports, targetList string
var simpleTCPScan, allPorts,check, filterClosed bool
var printLock sync.Mutex
var resultList []string


func CmdLineArgs() {
	flag.StringVar(&target, "t", "", "Single Target IPv4 Addresse or Hostname")
	flag.StringVar(&targetList, "iL", "", "List of Target IPv4 Adresses or Hostnames")
	flag.StringVar(&ports, "p", "", "Target Ports (comma-separated)")
	flag.BoolVar(&simpleTCPScan, "sT", false, "Start Simple TCP Scan")
	flag.BoolVar(&allPorts, "pA", false, "Scannt alle Ports")
	flag.BoolVar(&filterClosed, "fC", false, "Filtert die geschlossenen Ports heraus")

}

// überprüft die notwendigen Flags ((-p || -p-) && -t)
// -> bricht das Program ab, sollten sie nicht vorhanden sein
func CheckRequirements() {
	if target != "" {
	    target = strings.TrimSpace(target)
	    return
	}

	fmt.Println("[!] Benötigt werden Parameter Argumente zu target (Bspw. -t 1.1.1.1)")
	os.Exit(1)
}

// hier wird die PortRange initialisiert
// sollte allPorts gesetzt sein, wird die komplette port range erstellt
// sollte eine range, in der port deklarierung, angegeben sein, wird diese verarbeitet
func InitPortRange() []int {
	var floor, ceiling int
	var lports, limits []int
	
	// hier wird eine Portliste über die gesamte Portrange vorbereitet
	if allPorts {
		floor, ceiling = 1, 65535
		lports = make([]int, 0, ceiling-floor+1)
		for i := floor; i <= ceiling; i++ {
			lports = append(lports, i)
		}
	} else if ports == "" { // es wurde keine port range angegeben
		floor, ceiling = 1, 1000
		lports = make([]int, 0, ceiling-floor+1)
		for i := floor; i <= ceiling; i++ {
			lports = append(lports, i)
		}
	} else if strings.Contains(ports, "-") { // Voraussetzung; Eine Port Range (floor-ceiling) wurde als Argument für -p angegeben
		limitStr := strings.Split(ports, "-")

		if len(limitStr) != 2 {
		    fmt.Println("[!] Ungültige Portrange. Synthax: <floor-ceiling>")
		    os.Exit(1)
		}	

		// die limit Liste vorbereiten, indem aus den Strings Integer erstellt werden
		for _, p := range limitStr {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				fmt.Println("[!] Fehler bei den Port Ranges Limits")
				os.Exit(1)
			}
			limits = append(limits, n)
		}

		// hier wird die range gesetzt und anschließend die finale Portliste erstellt
		floor, ceiling = limits[0], limits[1]
		lports = make([]int, 0, ceiling-floor+1)
		for i := floor; i <= ceiling; i++ {
			lports = append(lports, i)
		}
	} else if strings.Contains(ports, ",") { 	// Voraussetzung; Komma-separierte Port Liste als Argument für -p
	// sollte eine komma-separierte liste an -p übergeben worden sein, wird diese augetrennt und jene Liste in eine Integer Liste umgewandelt 
		portStr := strings.Split(ports, ",")
		for _, p := range portStr {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				fmt.Println("[!] Fehler bei den Port Ranges Limits")
				os.Exit(1)
			}
			lports = append(lports, n)
		}

	} else { 	// Default State: Es wird ein einziger Port als Parameter übergeben 
	    n, err := strconv.Atoi(strings.TrimSpace(ports))
	    if err != nil {
	        fmt.Println("[!] Ungültiger Port")
	        os.Exit(1)
	    }
	    lports = append(lports, n)
	}


	return lports
}

// Hierüber können DNS Namensauflösungen durchgeführt werden 
/*func NsLookup() {
	// resolving multi targets #2DO

	// resolving single target 
	if target != "" {
		ips, err := net.LookupIP(target)
		if err != nil {
			panic(err)
			os.Exit(1)
		}
	}
}*/


// Simpler TCP Scan
func Worker(jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range jobs {
		addr := fmt.Sprintf("%s:%d", target, p)
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)

        printLock.Lock()

        if err != nil {
            if !filterClosed {
                result := fmt.Sprintf("closed %s", addr)
                fmt.Println(result)
                resultList = append(resultList, result)
            }
            printLock.Unlock()
            continue
        }

        result := fmt.Sprintf("open %s", addr)
        fmt.Println(result)
        resultList = append(resultList, result)
        conn.Close()

        printLock.Unlock()


	}


}

func SimpleTCPScan() {
	fmt.Println("[*] Starting GoMap Scan...")
	const workerCount = 100
	var wg sync.WaitGroup
	jobs := make(chan int, len(portList))
	
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go Worker(jobs, &wg)
	}

	for _, port := range portList {
		jobs <- port
	}

	close(jobs)

	wg.Wait()
}

func InitScan() {
	// Standard TCP Scan
	SimpleTCPScan()

}

func main() {
	// INITIALIZATION
	CmdLineArgs()
	flag.Parse()
	CheckRequirements()
	portList = InitPortRange()
	
	// MAIN PROGRAM LOGIC
	InitScan()
}
