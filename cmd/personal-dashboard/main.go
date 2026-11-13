// Personal dashboard companion: credentials stay on the machine displaying the dashboard.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:embed web/*
var web embed.FS

type settings struct {
	DeepSeek string `json:"deepseek_key"`
	OpenAI   string `json:"openai_admin_key"`
	CodexBar string `json:"codexbar_path"`
}
type metric struct {
	Name    string   `json:"name"`
	State   string   `json:"state"`
	Lines   []string `json:"lines"`
	Updated string   `json:"updated"`
}
type app struct {
	mu          sync.RWMutex
	config      settings
	metrics     []metric
	token, path string
	client      *http.Client
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	base := filepath.Dir(exe)
	if err = os.Chdir(base); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err = os.MkdirAll("personal-data", 0700); err != nil {
		log.Fatal(err)
	}
	a := &app{path: "personal-data/settings.json", client: &http.Client{Timeout: 20 * time.Second}}
	if b, e := os.ReadFile(a.path); e == nil {
		if json.Unmarshal(b, &a.config) != nil {
			log.Fatal("配置文件损坏，请恢复备份")
		}
	}
	if a.config.CodexBar == "" {
		a.config.CodexBar = filepath.Join(base, "CodexBarCLI")
	}
	seed := make([]byte, 32)
	if _, err = rand.Read(seed); err != nil {
		log.Fatal(err)
	}
	a.token = hex.EncodeToString(seed)
	a.metrics = []metric{{Name: "Codex / Claude", State: "loading", Lines: []string{"正在检查本机登录状态"}}, {Name: "DeepSeek API", State: "loading"}, {Name: "OpenAI API", State: "loading"}}
	listener, err := net.Listen("tcp", "127.0.0.1:8091")
	if err != nil {
		log.Fatal("配置服务端口 8091 无法使用: ", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.page)
	mux.HandleFunc("/api/settings", a.settings)
	mux.HandleFunc("/api/metrics", a.data)
	srv := &http.Server{Handler: a.guard(mux), ReadHeaderTimeout: 5 * time.Second}
	go srv.Serve(listener)
	go func() {
		for {
			a.refresh(ctx)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
			}
		}
	}()
	name := "glance"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	child := exec.CommandContext(ctx, filepath.Join(base, name), "--config", filepath.Join(base, "personal", "glance.yml"))
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err = child.Start(); err != nil {
		log.Fatal("无法启动 Glance: ", err)
	}
	fmt.Println("中文看板：http://127.0.0.1:8080　配置：http://127.0.0.1:8091　关闭此窗口停止服务")
	go func() {
		if err := child.Wait(); err != nil && ctx.Err() == nil {
			log.Print("Glance 已停止: ", err)
		}
		stop()
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	srv.Shutdown(shutdown)
}
func (a *app) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "127.0.0.1:8091" && r.Host != "localhost:8091" {
			http.Error(w, "主机不允许", 403)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
func (a *app) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	b, _ := web.ReadFile("web/index.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(strings.ReplaceAll(string(b), "__TOKEN__", a.token)))
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
func (a *app) settings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		a.mu.RLock()
		defer a.mu.RUnlock()
		writeJSON(w, map[string]any{"deepseek_configured": a.config.DeepSeek != "", "openai_configured": a.config.OpenAI != "", "codexbar_path": a.config.CodexBar})
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Setup-Token")), []byte(a.token)) != 1 {
		http.Error(w, "请重新打开配置页", 403)
		return
	}
	var s settings
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384)).Decode(&s) != nil {
		http.Error(w, "配置格式错误", 400)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	// Blank password fields preserve existing values; '-' explicitly removes a key.
	if s.DeepSeek == "" {
		s.DeepSeek = a.config.DeepSeek
	} else if s.DeepSeek == "-" {
		s.DeepSeek = ""
	}
	if s.OpenAI == "" {
		s.OpenAI = a.config.OpenAI
	} else if s.OpenAI == "-" {
		s.OpenAI = ""
	}
	s.CodexBar = strings.TrimSpace(s.CodexBar)
	b, _ := json.MarshalIndent(s, "", "  ")
	if err := save(a.path, b); err != nil {
		http.Error(w, "保存失败", 500)
		return
	}
	a.config = s
	writeJSON(w, map[string]string{"message": "已保存，最多 5 分钟后更新；也可以重启启动程序立即刷新。"})
}
func save(path string, b []byte) error {
	f, e := os.CreateTemp(filepath.Dir(path), ".settings-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(b)
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}
func (a *app) data(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	writeJSON(w, map[string]any{"metrics": a.metrics})
}
func (a *app) get(ctx context.Context, url, key string, out any) error {
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := a.client.Do(req)
	if e != nil {
		return errors.New("网络连接失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("平台返回 HTTP %d，请检查凭据和权限", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(out)
}
func (a *app) refresh(ctx context.Context) {
	a.mu.RLock()
	s := a.config
	a.mu.RUnlock()
	result := make([]metric, 3)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); result[0] = subscription(ctx, s.CodexBar) }()
	go func() { defer wg.Done(); result[1] = a.deepseek(ctx, s.DeepSeek) }()
	go func() { defer wg.Done(); result[2] = a.openai(ctx, s.OpenAI) }()
	wg.Wait()
	a.mu.Lock()
	a.metrics = result
	a.mu.Unlock()
}
func unset(name string) metric {
	return metric{Name: name, State: "unconfigured", Lines: []string{"未配置，请打开账号配置"}}
}
func failed(name string, e error) metric {
	return metric{Name: name, State: "error", Lines: []string{e.Error()}}
}
func ready(name string, lines []string) metric {
	return metric{Name: name, State: "ok", Lines: lines, Updated: time.Now().Format(time.RFC3339)}
}
func (a *app) deepseek(ctx context.Context, key string) metric {
	const name = "DeepSeek API 余额"
	if key == "" {
		return unset(name)
	}
	var v struct {
		Available bool `json:"is_available"`
		Balances  []struct {
			Currency string `json:"currency"`
			Total    string `json:"total_balance"`
			Granted  string `json:"granted_balance"`
			Paid     string `json:"topped_up_balance"`
		} `json:"balance_infos"`
	}
	if e := a.get(ctx, "https://api.deepseek.com/user/balance", key, &v); e != nil {
		return failed(name, e)
	}
	if len(v.Balances) == 0 {
		return failed(name, errors.New("平台未返回余额"))
	}
	lines := []string{}
	for _, b := range v.Balances {
		if b.Total == "" || b.Currency == "" {
			return failed(name, errors.New("余额字段缺失"))
		}
		lines = append(lines, "可用余额："+b.Total+" "+b.Currency, "充值："+b.Paid+" · 赠金："+b.Granted)
	}
	if !v.Available {
		lines = append(lines, "当前余额不可用于调用")
	}
	return ready(name, lines)
}
func (a *app) openai(ctx context.Context, key string) metric {
	const name = "OpenAI API 本月费用"
	if key == "" {
		return unset(name)
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	endpoint := fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&bucket_width=1d&limit=31", start)
	total := 0.0
	for page := 0; page < 100; page++ {
		var v struct {
			Data []struct {
				Results []struct {
					Amount *struct {
						Value    float64 `json:"value"`
						Currency string  `json:"currency"`
					} `json:"amount"`
				} `json:"results"`
			} `json:"data"`
			More bool   `json:"has_more"`
			Next string `json:"next_page"`
		}
		if e := a.get(ctx, endpoint, key, &v); e != nil {
			return failed(name, e)
		}
		if v.Data == nil {
			return failed(name, errors.New("平台未返回费用数据"))
		}
		for _, bucket := range v.Data {
			for _, r := range bucket.Results {
				if r.Amount == nil || r.Amount.Currency != "usd" {
					return failed(name, errors.New("费用数据缺失或币种不支持"))
				}
				total += r.Amount.Value
			}
		}
		if !v.More {
			return ready(name, []string{fmt.Sprintf("本月已消费：$%.2f（UTC 月份）", total), "真实余额：请查看 OpenAI 官方账单页"})
		}
		if v.Next == "" {
			return failed(name, errors.New("费用分页信息缺失"))
		}
		endpoint = fmt.Sprintf("https://api.openai.com/v1/organization/costs?start_time=%d&bucket_width=1d&limit=31&page=%s", start, url.QueryEscape(v.Next))
	}
	return failed(name, errors.New("费用分页过多，未显示不完整金额"))
}
func subscription(ctx context.Context, path string) metric {
	const name = "Codex / Claude 订阅额度"
	if path == "" {
		return unset(name)
	}
	if _, e := os.Stat(path); e != nil {
		return failed(name, errors.New("尚未安装 CodexBar，请查看配置页的安装说明"))
	}
	c, cancel := context.WithTimeout(ctx, 70*time.Second)
	defer cancel()
	cmd := exec.CommandContext(c, path, "dashboard", "--identity", "redacted", "--timeout", "60")
	// Never render provider diagnostics, which may contain account or credential details.
	b, e := cmd.Output()
	if e != nil {
		return failed(name, errors.New("额度采集失败，请在本机完成 Codex / Claude 登录"))
	}
	return parseSubscriptions(b)
}
func parseSubscriptions(b []byte) metric {
	const name = "Codex / Claude 订阅额度"
	var v struct {
		Providers []struct {
			ID      string `json:"id"`
			Error   any    `json:"error"`
			Updated string `json:"updatedAt"`
			Windows []struct {
				Kind      string   `json:"kind"`
				Remaining *float64 `json:"remainingPercent"`
				Reset     string   `json:"resetAt"`
			} `json:"windows"`
		} `json:"providers"`
	}
	if json.Unmarshal(b, &v) != nil {
		return failed(name, errors.New("额度数据格式错误"))
	}
	lines := []string{}
	for _, p := range v.Providers {
		if p.ID != "codex" && p.ID != "claude" {
			continue
		}
		label := map[string]string{"codex": "Codex", "claude": "Claude"}[p.ID]
		if p.Error != nil {
			lines = append(lines, label+"：读取失败，请检查登录")
			continue
		}
		t, e := time.Parse(time.RFC3339, p.Updated)
		if e != nil || time.Since(t) > 15*time.Minute {
			lines = append(lines, label+"：数据已过期，请检查采集器")
			continue
		}
		if len(p.Windows) == 0 {
			lines = append(lines, label+"：暂无额度窗口")
			continue
		}
		for _, w := range p.Windows {
			if w.Remaining == nil {
				lines = append(lines, label+"：余量未知")
				continue
			}
			window := map[string]string{"session": "当前周期", "weekly": "本周"}[w.Kind]
			if window == "" {
				window = "其他额度"
			}
			line := fmt.Sprintf("%s · %s剩余 %.0f%%", label, window, *w.Remaining)
			if rt, e := time.Parse(time.RFC3339, w.Reset); e == nil {
				line += " · " + rt.Local().Format("01-02 15:04") + " 重置"
			}
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return unset(name)
	}
	return ready(name, lines)
}
