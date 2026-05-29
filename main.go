package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"radar/resolvers"
	"radar/resolvers/simple"
	"sync"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

var debug bool

type ResolutionResult struct {
	Name string
	IP   string
	Err  error
}

type CheckResult struct {
	Name      string
	Status    string
	Message   string
	Durations time.Duration
}

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
}

func main() {
	flag.Parse()

	if debug {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		slog.SetDefault(logger)
		slog.Debug("Debug mode enabled")
	}

	client := http.Client{
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
				Resolver: &net.Resolver{
					PreferGo: true,
				},
			}).DialContext,
		},
	}

	ctx := context.Background()

	fmt.Println(Green + "██████╗  █████╗ ██████╗  █████╗ ██████╗")
	fmt.Println("██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔══██╗")
	fmt.Println("██████╔╝███████║██║  ██║███████║██████╔╝")
	fmt.Println("██╔══██╗██╔══██║██║  ██║██╔══██║██╔══██╗")
	fmt.Println("██║  ██║██║  ██║██████╔╝██║  ██║██║  ██║")
	fmt.Println("╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝")
	fmt.Println("by Deerain\n" + Reset)

	activeResolvers := map[string]resolvers.Resolver{
		"ipify":    simple.New("https://api.ipify.org", client),
		"myip.com": simple.New("https://api.myip.com", client),
		"2ip.io":   simple.New("https://api.2ip.io", client),
		"beget.ru": simple.New("https://ip.beget.ru", client),
	}

	blockedServices := map[string]string{
		"Instagram": "https://instagram.com",
		"Twitter/X": "https://x.com",
		"Facebook":  "https://facebook.com",
		"OpenAI":    "https://openai.com",
		"Telegram":  "https://telegram.org",
		"Youtube":   "https://youtube.com",
	}

	ipChan := make(chan ResolutionResult, len(activeResolvers))
	checkChan := make(chan CheckResult, len(blockedServices))

	var wgResolvers sync.WaitGroup
	var wgCheckers sync.WaitGroup

	fmt.Printf("%s[+] Running diagnostic benchmarks...%s\n\n", Blue, Reset)

	startTime := time.Now()

	go StartIPResolvers(activeResolvers, &wgResolvers, ctx, ipChan)
	go StartBlockChecker(blockedServices, &wgCheckers, ctx, &client, checkChan)

	var fetchedIPs = make([]ResolutionResult, 0, len(activeResolvers))
	for i := 0; i < len(activeResolvers); i++ {
		fetchedIPs = append(fetchedIPs, <-ipChan)
	}

	var checkedServices = make([]CheckResult, 0, len(blockedServices))
	for i := 0; i < len(blockedServices); i++ {
		checkedServices = append(checkedServices, <-checkChan)
	}

	fmt.Printf("%s--- IP Addresses Route ---%s\n", Purple, Reset)
	for _, res := range fetchedIPs {
		if res.Err != nil {
			fmt.Printf("Failed to get ip address from %s%s%s: %v\n", Red, res.Name, Reset, res.Err)
		} else {
			fmt.Printf("IP Address: %s%-15s%s (%s)\n", Green, res.IP, Reset, res.Name)
		}
	}
	if HasSplitTunneling(fetchedIPs) {
		fmt.Printf("\n%s[ WARN ]%s Split tunneling detected\n\n", Yellow, Reset)
	} else {
		fmt.Printf("\n%s[ INFO ]%s Single gateway routing detected\n\n", Blue, Reset)
	}

	fmt.Printf("%s--- Blocked Services ---%s\n", Purple, Reset)
	for _, res := range checkedServices {
		switch res.Status {
		case "OK":
			fmt.Printf("%s[ OK ]%s   %-12s: %s\n", Green, Reset, res.Name, res.Message)
		case "FAIL":
			fmt.Printf("%s[ FAIL ]%s %-12s: %s\n", Red, Reset, res.Name, res.Message)
		case "WARN":
			fmt.Printf("%s[ WARN ]%s %-12s: %s\n", Yellow, Reset, res.Name, res.Message)
		}
	}

	fmt.Printf("\nDone with: %v\n", time.Since(startTime))
}

func StartIPResolvers(activeResolvers map[string]resolvers.Resolver, wg *sync.WaitGroup, ctx context.Context, ch chan<- ResolutionResult) {
	for key, resolver := range activeResolvers {
		wg.Add(1)
		go func(k string, r resolvers.Resolver) {
			defer wg.Done()
			ip, err := r.Resolve(ctx)
			ch <- ResolutionResult{
				Name: k,
				IP:   ip,
				Err:  err,
			}
		}(key, resolver)
	}

	wg.Wait()
	close(ch)
}

func StartBlockChecker(services map[string]string, wg *sync.WaitGroup, ctx context.Context, client *http.Client, ch chan<- CheckResult) {
	for name, url := range services {
		wg.Add(1)

		go func(n string, u string) {
			defer wg.Done()
			ch <- CheckBlockedService(ctx, n, u, client)
		}(name, url)

	}

	wg.Wait()
	close(ch)
}

func HasSplitTunneling(ips []ResolutionResult) bool {
	var validIPs []ResolutionResult
	for _, ip := range ips {
		if ip.Err == nil && ip.IP != "" {
			validIPs = append(validIPs, ip)
		}
	}

	if len(validIPs) < 2 {
		return false
	}

	firstIP := validIPs[0]
	for _, res := range validIPs[1:] {
		slog.Debug("Check ip", "first", firstIP.IP, "check ip", res.IP)
		if res.IP != firstIP.IP {
			return true
		}
	}
	return false
}

func CheckBlockedService(ctx context.Context, name string, url string, client *http.Client) CheckResult {
	subCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(subCtx, "HEAD", url, nil)
	if err != nil {
		return CheckResult{Name: name, Status: "FAIL", Message: "Failed to create request"}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	startPing := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startPing).Round(time.Microsecond)

	if err != nil {
		return CheckResult{
			Name:    name,
			Status:  "FAIL",
			Message: fmt.Sprintf("Unavailable (Timeout/Block) | %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return CheckResult{
			Name:    name,
			Status:  "OK",
			Message: fmt.Sprintf("Available | Response: %d | Time: %v", resp.StatusCode, duration),
		}
	}

	slog.Debug("Request", "url", url, "status", resp.StatusCode, "data", resp.Body)

	return CheckResult{
		Name:    name,
		Status:  "WARN",
		Message: fmt.Sprintf("Status code: %d", resp.StatusCode),
	}
}
