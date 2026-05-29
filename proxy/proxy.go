package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type ProxyInfo struct {
	Detected bool
	Type     string
	Address  string
	URL      *url.URL
}

// DetectActiveProxies ищет прокси и парсит его адрес в структуру URL
func DetectActiveProxies() ProxyInfo {
	// 1. Проверяем системные переменные окружения
	envProxies := []string{"ALL_PROXY", "all_proxy", "HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"}
	for _, env := range envProxies {
		if val := os.Getenv(env); val != "" {
			// Дописываем схему, если пользователь ввел просто ip:port
			rawURL := val
			if !strings.Contains(rawURL, "://") {
				rawURL = "http://" + rawURL
			}

			parsedURL, err := url.Parse(rawURL)
			if err == nil {
				pType := "HTTP"
				if strings.HasPrefix(parsedURL.Scheme, "socks") {
					pType = "SOCKS5"
				}
				return ProxyInfo{Detected: true, Type: "System (" + pType + ")", Address: val, URL: parsedURL}
			}
		}
	}

	// 2. Список дефолтных портов локальных прокси-клиентов
	commonPorts := []struct {
		addr   string
		scheme string
		name   string
	}{
		{"127.0.0.1:1080", "socks5", "SOCKS5 (Shadowsocks/Xray/Sing-box)"},
		{"127.0.0.1:10808", "socks5", "SOCKS5 (Xray Windows)"},
		{"127.0.0.1:7890", "socks5", "SOCKS5/HTTP (Clash/Mihomo)"},
		{"127.0.0.1:2080", "socks5", "SOCKS5 (Nekoray)"},
		{"127.0.0.1:9050", "socks5", "SOCKS5 (Tor)"},
	}

	for _, p := range commonPorts {
		d := net.Dialer{Timeout: 40 * time.Millisecond}
		conn, err := d.DialContext(context.Background(), "tcp", p.addr)
		if err == nil {
			conn.Close()
			parsedURL, _ := url.Parse(fmt.Sprintf("%s://%s", p.scheme, p.addr))
			return ProxyInfo{Detected: true, Type: p.name, Address: p.addr, URL: parsedURL}
		}
	}

	return ProxyInfo{Detected: false}
}

// ConfigureHttpClient настраивает глобальный транспорт в зависимости от типа прокси
func ConfigureHttpClient(httpClient *http.Client, p ProxyInfo) {
	if !p.Detected || p.URL == nil {
		// Если прокси не найден, оставляем дефолтный чистый Go-транспорт
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment, // на всякий случай
		}
		return
	}

	// Если это SOCKS5
	if strings.HasPrefix(p.URL.Scheme, "socks") {
		dialer, err := proxy.FromURL(p.URL, proxy.Direct)
		if err == nil {
			// Кастуем контекстный диалер для правильной работы таймаутов
			if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
				httpClient.Transport = &http.Transport{
					DialContext: contextDialer.DialContext,
				}
				return
			}
		}
	}

	// Если это стандартный HTTP-прокси
	httpClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(p.URL),
	}
}
