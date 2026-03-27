package integration

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"bloop-tunnel/internal/config"
	"bloop-tunnel/internal/logging"
	"bloop-tunnel/internal/relay"
	"bloop-tunnel/internal/relay/registry"
	"bloop-tunnel/internal/relay/routing"
	relaysession "bloop-tunnel/internal/relay/session"
)

var (
	stackOnce sync.Once
	stackErr  error
	stackEnv  *testStack
)

type testStack struct {
	mu               sync.Mutex
	root             string
	echoAddr         string
	relayAddr        string
	relayConfigPath  string
	clientConfigPath string
	echoCmd          *exec.Cmd
	clientCmd        *exec.Cmd
	clientLogMu      sync.Mutex
	clientLogLines   []string
	relayConfig      *config.RelayConfig
	relayLogger      *slog.Logger
	relayRegistry    *registry.Registry
	relayManager     *relaysession.Manager
	relayServer      *http.Server
}

func TestMain(m *testing.M) {
	if err := startSharedTestStack(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	cleanupSharedTestStack()
	os.Exit(code)
}

func startSharedTestStack() error {
	stackOnce.Do(func() {
		root, err := projectRoot()
		if err != nil {
			stackErr = err
			return
		}

		echoAddr, err := reserveLocalAddr()
		if err != nil {
			stackErr = fmt.Errorf("reserve echo addr: %w", err)
			return
		}
		relayAddr, err := reserveLocalAddr()
		if err != nil {
			stackErr = fmt.Errorf("reserve relay addr: %w", err)
			return
		}

		stack := &testStack{root: root, echoAddr: echoAddr, relayAddr: relayAddr}

		relayConfigPath, err := writeRelayConfig(root, relayAddr)
		if err != nil {
			stackErr = err
			return
		}
		stack.relayConfigPath = relayConfigPath

		relayCfg, err := config.LoadRelayConfig(relayConfigPath)
		if err != nil {
			_ = os.Remove(relayConfigPath)
			stackErr = fmt.Errorf("load relay config: %w", err)
			return
		}
		stack.relayConfig = relayCfg
		stack.relayLogger = logging.New(relayCfg.Logging.Level)
		stack.relayRegistry = registry.New()
		stack.relayManager = relaysession.NewManager()

		clientConfigPath, err := writeClientConfig(root, relayAddr, echoAddr)
		if err != nil {
			_ = os.Remove(relayConfigPath)
			stackErr = err
			return
		}
		stack.clientConfigPath = clientConfigPath

		if err := stack.startEchoLocked(); err != nil {
			cleanupStack(stack)
			stackErr = err
			return
		}
		if err := stack.startRelayLocked(); err != nil {
			cleanupStack(stack)
			stackErr = err
			return
		}
		if err := stack.startClientLocked(); err != nil {
			cleanupStack(stack)
			stackErr = err
			return
		}

		stackEnv = stack
	})

	return stackErr
}

func cleanupSharedTestStack() {
	if stackEnv != nil {
		cleanupStack(stackEnv)
	}
}

func cleanupStack(stack *testStack) {
	if stack == nil {
		return
	}
	stack.mu.Lock()
	defer stack.mu.Unlock()
	killProcess(stack.clientCmd)
	stack.clientCmd = nil
	if stack.relayServer != nil {
		_ = stack.relayServer.Close()
		stack.relayServer = nil
	}
	killProcess(stack.echoCmd)
	stack.echoCmd = nil
	if stack.clientConfigPath != "" {
		_ = os.Remove(stack.clientConfigPath)
	}
	if stack.relayConfigPath != "" {
		_ = os.Remove(stack.relayConfigPath)
	}
}

func (s *testStack) restartRelay() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.relayServer != nil {
		_ = s.relayServer.Close()
		s.relayServer = nil
	}
	s.relayRegistry = registry.New()
	s.relayManager = relaysession.NewManager()
	return s.startRelayLocked()
}

func (s *testStack) disconnectClientTransport() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.relayManager.CloseAny()
}

func (s *testStack) waitForClientLog(needle string) error {
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		s.clientLogMu.Lock()
		lines := append([]string(nil), s.clientLogLines...)
		s.clientLogMu.Unlock()
		for _, line := range lines {
			if strings.Contains(line, needle) {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("client log line not observed: %s", needle)
}

func (s *testStack) killClient() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	killProcess(s.clientCmd)
	s.clientCmd = nil
	return nil
}

func (s *testStack) restartClient() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	killProcess(s.clientCmd)
	s.clientCmd = nil
	return s.startClientLocked()
}

func (s *testStack) startEchoLocked() error {
	echoCmd := exec.Command("python3", "-c", echoServerProgram(s.echoAddr))
	echoCmd.Stdout = os.Stdout
	echoCmd.Stderr = os.Stderr
	if err := echoCmd.Start(); err != nil {
		return fmt.Errorf("start echo server: %w", err)
	}
	s.echoCmd = echoCmd
	waitForPort(s.echoAddr)
	return nil
}

func (s *testStack) startRelayLocked() error {
	router := routing.New(s.relayRegistry)
	proxy := relay.NewRequestProxy(s.relayManager, s.relayLogger)
	sessionHandler := relaysession.NewHandler(s.relayConfig, s.relayManager, s.relayRegistry, s.relayLogger)
	sessionHandler.OnEnvelope = proxy.HandleEnvelope
	httpHandler := &relay.HTTPHandler{Router: router, Proxy: proxy}

	mux := http.NewServeMux()
	mux.Handle("/connect", sessionHandler)
	mux.Handle("/", httpHandler)

	server := &http.Server{Addr: s.relayAddr, Handler: mux}
	s.relayServer = server
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.relayLogger.Error("relay server failed", "error", err)
		}
	}()
	waitForPort(s.relayAddr)
	return nil
}

func (s *testStack) startClientLocked() error {
	clientCmd := exec.Command("go", "run", "./cmd/bloop-client", "--config", s.clientConfigPath)
	clientCmd.Dir = s.root
	clientStdout, err := clientCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("client stdout pipe: %w", err)
	}
	clientCmd.Stderr = clientCmd.Stdout
	if err := clientCmd.Start(); err != nil {
		return fmt.Errorf("start client: %w", err)
	}
	s.clientCmd = clientCmd
	s.clientLogLines = nil
	lineCh := streamLogs(clientStdout, func(line string) {
		s.clientLogMu.Lock()
		defer s.clientLogMu.Unlock()
		s.clientLogLines = append(s.clientLogLines, line)
		if len(s.clientLogLines) > 400 {
			s.clientLogLines = s.clientLogLines[len(s.clientLogLines)-400:]
		}
	})
	if err := waitForLogLines(lineCh, "client registered tunnels successfully"); err != nil {
		killProcess(clientCmd)
		s.clientCmd = nil
		return err
	}
	return nil
}

func assertEventuallyReachable(t *testing.T, method, host, path string, headers map[string]string, body io.Reader, wantStatus int, wantSubstring string) {
	t.Helper()
	deadline := time.Now().Add(12 * time.Second)
	var requestBody string
	if body != nil {
		data, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		requestBody = string(data)
	}

	var lastErr error
	var lastStatus int
	var lastBody string

	for time.Now().Before(deadline) {
		var reqBody io.Reader
		if requestBody != "" {
			reqBody = strings.NewReader(requestBody)
		}
		req, err := http.NewRequestWithContext(t.Context(), method, "http://"+stackEnv.relayAddr+path, reqBody)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Host = host
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(300 * time.Millisecond)
			continue
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode
		lastBody = string(bodyBytes)
		if resp.StatusCode == wantStatus && strings.Contains(lastBody, wantSubstring) {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}

	if lastErr != nil {
		t.Fatalf("request never recovered: %v", lastErr)
	}
	t.Fatalf("request never matched expected status/body; got status=%d body=%q", lastStatus, lastBody)
}

func projectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(wd, "../..")), nil
}

func reserveLocalAddr() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		return "", err
	}
	return addr, nil
}

func portComponent(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	return port
}

func writeRelayConfig(root, relayAddr string) (string, error) {
	content := strings.Join([]string{
		"domain: bloop.to",
		fmt.Sprintf("listen_addr: \":%s\"", portComponent(relayAddr)),
		"trusted_proxies:",
		"  - 127.0.0.1/32",
		"client_tokens:",
		"  - name: laptop-main",
		"    token: test-token",
		"hostname_generation:",
		"  mode: random",
		"  length: 8",
		"logging:",
		"  level: debug",
		"  format: json",
	}, "\n") + "\n"

	path := filepath.Join(root, "test/integration/.tmp-relay.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("write relay config: %w", err)
	}
	return path, nil
}

func writeClientConfig(root, relayAddr, echoAddr string) (string, error) {
	content := strings.Join([]string{
		fmt.Sprintf("relay_url: ws://%s/connect", relayAddr),
		"auth_token: test-token",
		"reconnect:",
		"  initial_delay_ms: 1000",
		"  max_delay_ms: 30000",
		"logging:",
		"  level: debug",
		"  format: json",
		"tunnels:",
		"  - name: public-app",
		"    hostname: public.bloop.to",
		fmt.Sprintf("    local_addr: %s", echoAddr),
		"    access: public",
		"  - name: basic-app",
		"    hostname: basic.bloop.to",
		fmt.Sprintf("    local_addr: %s", echoAddr),
		"    access: basic_auth",
		"    basic_auth:",
		"      username: gene",
		"      password: secretpass",
		"  - name: token-app",
		"    hostname: token.bloop.to",
		fmt.Sprintf("    local_addr: %s", echoAddr),
		"    access: token_protected",
		"    token: topsecret",
	}, "\n") + "\n"

	path := filepath.Join(root, "test/integration/.tmp-client.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("write client config: %w", err)
	}
	return path, nil
}

func echoServerProgram(echoAddr string) string {
	return strings.Join([]string{
		"from http.server import BaseHTTPRequestHandler, HTTPServer",
		"class Handler(BaseHTTPRequestHandler):",
		"    def _write_response(self, payload):",
		"        self.send_response(200)",
		"        self.send_header('Content-Type', 'text/plain; charset=utf-8')",
		"        self.send_header('Content-Length', str(len(payload)))",
		"        self.end_headers()",
		"        self.wfile.write(payload)",
		"    def do_GET(self):",
		"        body = f'echo GET {self.path}\\n'.encode()",
		"        self._write_response(body)",
		"    def do_POST(self):",
		"        length = int(self.headers.get('Content-Length', '0'))",
		"        body = self.rfile.read(length)",
		"        response = b'echo POST ' + self.path.encode() + b'\\n' + body",
		"        self._write_response(response)",
		"    def log_message(self, format, *args):",
		"        pass",
		fmt.Sprintf("HTTPServer(('127.0.0.1', %s), Handler).serve_forever()", portComponent(echoAddr)),
	}, "\n")
}

func waitForPort(addr string) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	panic(fmt.Sprintf("port never became ready: %s", addr))
}

type logResult struct {
	line string
	err  error
}

func streamLogs(r io.Reader, onLine func(string)) <-chan logResult {
	lines := make(chan logResult, 64)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if onLine != nil {
				onLine(line)
			}
			lines <- logResult{line: line}
		}
		if err := scanner.Err(); err != nil {
			lines <- logResult{err: err}
		}
		close(lines)
	}()
	return lines
}

func waitForLog(r io.Reader, needle string) error {
	return waitForLogLines(streamLogs(r, nil), needle)
}

func waitForLogLines(lines <-chan logResult, needle string) error {
	deadline := time.NewTimer(12 * time.Second)
	defer deadline.Stop()

	for {
		select {
		case item, ok := <-lines:
			if !ok {
				return fmt.Errorf("log line not observed before stream ended: %s", needle)
			}
			if item.err != nil {
				return fmt.Errorf("scanner error: %w", item.err)
			}
			if strings.Contains(item.line, needle) {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("log line not observed: %s", needle)
		}
	}
}

func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}
