package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	PrinterPath    string `json:"printerPath"`
	COHCCOM        string `json:"cohcCom"`
	SoundCOM       string `json:"soundCom"`
	COHCMode       string `json:"cohcMode"`
	SoundMode      string `json:"soundMode"`
	COHCUDPIP      string `json:"cohcUdpIp"`
	COHCUDPPort    int    `json:"cohcUdpPort"`
	COHCLocalPort  int    `json:"cohcLocalPort"`
	SoundUDPIP     string `json:"soundUdpIp"`
	SoundUDPPort   int    `json:"soundUdpPort"`
	SoundLocalPort int    `json:"soundLocalPort"`
	UDPTimeoutMs   int    `json:"udpTimeoutMs"`
	BaudRate       int    `json:"baudRate"`
	ReadInterval   int    `json:"readInterval"`
	TemplateName   string `json:"templateName"`
	CaptureName    string `json:"captureName"`
	BrakeForceUnit string `json:"brakeForceUnit"`
	WeightUnit     string `json:"weightUnit"`
}

var configFile = "setting.conf"
var cfg = Config{
	PrinterPath:    "printerpath",
	COHCCOM:        "COM1",
	SoundCOM:       "COM2",
	COHCMode:       "COM",
	SoundMode:      "COM",
	COHCUDPIP:      "192.168.1.145",
	COHCUDPPort:    7795,
	COHCLocalPort:  2002,
	SoundUDPIP:     "192.168.1.145",
	SoundUDPPort:   7795,
	SoundLocalPort: 2002,
	UDPTimeoutMs:   2000,
	BaudRate:       9600,
	ReadInterval:   2,
	TemplateName:   "CalibrateValue.tif",
	CaptureName:    "capture.tif",
	BrakeForceUnit: "N",
	WeightUnit:     "kgx10",
}

func main() {
	loadConfig()
	os.MkdirAll(cfg.PrinterPath, 0755)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/files", handleFiles)
	mux.HandleFunc("/api/image", handleImage)
	mux.HandleFunc("/api/upload", handleUpload)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/cohc/read", handleCOHC)
	mux.HandleFunc("/api/sound/read", handleSound)
	mux.HandleFunc("/api/udp/test", handleUDPTest)
	mux.HandleFunc("/api/vehicle/detect", handleVehicleDetect)
	mux.HandleFunc("/api/vehicle/yolo-status", handleYoloStatus)

	url := "http://127.0.0.1:8080"
	go func() { time.Sleep(700 * time.Millisecond); openBrowser(url) }()
	log.Println("InspectionAgent V17 started at", url)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}

func loadConfig() {
	b, err := os.ReadFile(configFile)
	if err == nil {
		_ = json.Unmarshal(b, &cfg)
	}
	if cfg.PrinterPath == "" {
		cfg.PrinterPath = "printerpath"
	}
	if cfg.COHCMode == "" {
		cfg.COHCMode = "COM"
	}
	if cfg.SoundMode == "" {
		cfg.SoundMode = "COM"
	}
	if cfg.COHCUDPIP == "" {
		cfg.COHCUDPIP = "192.168.1.145"
	}
	if cfg.COHCUDPPort == 0 {
		cfg.COHCUDPPort = 7795
	}
	if cfg.COHCLocalPort == 0 {
		cfg.COHCLocalPort = 2002
	}
	if cfg.SoundUDPIP == "" {
		cfg.SoundUDPIP = "192.168.1.145"
	}
	if cfg.SoundUDPPort == 0 {
		cfg.SoundUDPPort = 7795
	}
	if cfg.SoundLocalPort == 0 {
		cfg.SoundLocalPort = 2002
	}
	if cfg.UDPTimeoutMs == 0 {
		cfg.UDPTimeoutMs = 2000
	}
	if cfg.BaudRate == 0 {
		cfg.BaudRate = 9600
	}
	if cfg.ReadInterval < 1 || cfg.ReadInterval > 5 {
		cfg.ReadInterval = 2
	}
	if cfg.TemplateName == "" {
		cfg.TemplateName = "CalibrateValue.tif"
	}
	if cfg.CaptureName == "" {
		cfg.CaptureName = "capture.tif"
	}
}
func saveConfig() {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(configFile, b, 0644)
	os.MkdirAll(cfg.PrinterPath, 0755)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, cfg)
		return
	}
	if r.Method == http.MethodPost {
		var n Config
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if n.PrinterPath != "" {
			cfg.PrinterPath = n.PrinterPath
		}
		if n.COHCCOM != "" {
			cfg.COHCCOM = n.COHCCOM
		}
		if n.SoundCOM != "" {
			cfg.SoundCOM = n.SoundCOM
		}
		if n.COHCMode != "" {
			cfg.COHCMode = strings.ToUpper(n.COHCMode)
		}
		if n.SoundMode != "" {
			cfg.SoundMode = strings.ToUpper(n.SoundMode)
		}
		if n.COHCUDPIP != "" {
			cfg.COHCUDPIP = n.COHCUDPIP
		}
		if n.COHCUDPPort > 0 {
			cfg.COHCUDPPort = n.COHCUDPPort
		}
		if n.COHCLocalPort > 0 {
			cfg.COHCLocalPort = n.COHCLocalPort
		}
		if n.SoundUDPIP != "" {
			cfg.SoundUDPIP = n.SoundUDPIP
		}
		if n.SoundUDPPort > 0 {
			cfg.SoundUDPPort = n.SoundUDPPort
		}
		if n.SoundLocalPort > 0 {
			cfg.SoundLocalPort = n.SoundLocalPort
		}
		if n.UDPTimeoutMs > 0 {
			cfg.UDPTimeoutMs = n.UDPTimeoutMs
		}
		if n.BaudRate > 0 {
			cfg.BaudRate = n.BaudRate
		}
		if n.ReadInterval >= 1 && n.ReadInterval <= 5 {
			cfg.ReadInterval = n.ReadInterval
		}
		if n.TemplateName != "" {
			cfg.TemplateName = n.TemplateName
		}
		if n.CaptureName != "" {
			cfg.CaptureName = n.CaptureName
		}
		saveConfig()
		writeJSON(w, map[string]any{"ok": true, "config": cfg})
		return
	}
	http.Error(w, "method not allowed", 405)
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	os.MkdirAll(cfg.PrinterPath, 0755)
	entries, _ := os.ReadDir(cfg.PrinterPath)
	var files []map[string]any
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".tif") || strings.HasSuffix(lower, ".tiff") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
			info, _ := e.Info()
			files = append(files, map[string]any{"name": name, "size": info.Size(), "modified": info.ModTime().Format("2006-01-02 15:04:05")})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i]["modified"].(string) > files[j]["modified"].(string) })
	writeJSON(w, map[string]any{
		"printerPath":    cfg.PrinterPath,
		"templateName":   cfg.TemplateName,
		"captureName":    cfg.CaptureName,
		"templateExists": fileExists(filepath.Join(cfg.PrinterPath, cfg.TemplateName)),
		"captureExists":  fileExists(filepath.Join(cfg.PrinterPath, cfg.CaptureName)),
		"files":          files,
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"app":            "InspectionAgent V17",
		"time":           time.Now().Format("2006-01-02 15:04:05"),
		"templateExists": fileExists(filepath.Join(cfg.PrinterPath, cfg.TemplateName)),
		"captureExists":  fileExists(filepath.Join(cfg.PrinterPath, cfg.CaptureName)),
	})
}

func handleCOHC(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(cfg.COHCMode) == "UDP" {
		raw, err := udpRequest(cfg.COHCUDPIP, cfg.COHCUDPPort, cfg.COHCLocalPort, "e", time.Duration(cfg.UDPTimeoutMs)*time.Millisecond)
		co, hc := parseCOHC(raw)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "mode": "UDP", "error": err.Error(), "raw": raw, "co": co, "hc": hc})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "mode": "UDP", "ip": cfg.COHCUDPIP, "remotePort": cfg.COHCUDPPort, "localPort": cfg.COHCLocalPort, "command": "e", "raw": raw, "co": co, "hc": hc})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "co": "0.12", "hc": "112", "raw": "CO : 0.12% HC : 112 ppm", "com": cfg.COHCCOM, "mode": "COM", "demo": true, "note": "ยังเป็น demo COM: เพิ่ม serial parser ตามรุ่นเครื่องมือได้"})
}
func handleSound(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(cfg.SoundMode) == "UDP" {
		raw, err := udpRequest(cfg.SoundUDPIP, cfg.SoundUDPPort, cfg.SoundLocalPort, "a", time.Duration(cfg.UDPTimeoutMs)*time.Millisecond)
		sound := parseFirstNumber(raw)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "mode": "UDP", "error": err.Error(), "raw": raw, "sound": sound, "unit": "dB"})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "mode": "UDP", "ip": cfg.SoundUDPIP, "remotePort": cfg.SoundUDPPort, "localPort": cfg.SoundLocalPort, "command": "a", "raw": raw, "sound": sound, "unit": "dB"})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "sound": "82.1", "unit": "dB", "com": cfg.SoundCOM, "mode": "COM", "demo": true, "note": "ยังเป็น demo COM: เพิ่ม serial parser ตามรุ่นเครื่องมือได้"})
}

func handleUDPTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		LocalPort int    `json:"localPort"`
		Command   string `json:"command"`
		TimeoutMs int    `json:"timeoutMs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = cfg.UDPTimeoutMs
	}
	raw, err := udpRequest(req.IP, req.Port, req.LocalPort, req.Command, time.Duration(req.TimeoutMs)*time.Millisecond)
	res := map[string]any{"ok": err == nil, "ip": req.IP, "port": req.Port, "localPort": req.LocalPort, "command": req.Command, "raw": raw, "co": "", "hc": "", "sound": parseFirstNumber(raw)}
	co, hc := parseCOHC(raw)
	res["co"], res["hc"] = co, hc
	if err != nil {
		res["error"] = err.Error()
	}
	writeJSON(w, res)
}

func udpRequest(ip string, remotePort, localPort int, command string, timeout time.Duration) (string, error) {
	if ip == "" || remotePort <= 0 {
		return "", fmt.Errorf("กรุณาระบุ IP และ remote port")
	}
	if command == "" {
		return "", fmt.Errorf("กรุณาระบุ command")
	}
	localAddr := &net.UDPAddr{IP: net.IPv4zero, Port: localPort}
	remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, remotePort))
	if err != nil {
		return "", err
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.WriteToUDP([]byte(command), remoteAddr)
	if err != nil {
		return "", err
	}
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 4096)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buf[:n])) + "  [from " + addr.String() + "]", nil
}

func parseCOHC(raw string) (string, string) {
	clean := strings.ReplaceAll(raw, "  ", " ")
	co, hc := "", ""
	upper := strings.ToUpper(clean)
	if i := strings.Index(upper, "CO"); i >= 0 {
		part := clean[i:]
		if j := strings.Index(part, "%"); j >= 0 {
			co = parseFirstNumber(part[:j])
		}
	}
	if i := strings.Index(upper, "HC"); i >= 0 {
		part := clean[i:]
		if j := strings.Index(strings.ToUpper(part), "PPM"); j >= 0 {
			hc = parseFirstNumber(part[:j])
		} else {
			hc = parseFirstNumber(part)
		}
	}
	return co, hc
}

func parseFirstNumber(s string) string {
	var b strings.Builder
	started := false
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
			started = true
		} else if started {
			break
		}
	}
	return b.String()
}

func handleYoloStatus(w http.ResponseWriter, r *http.Request) {
	py := findPython()
	if py == "" {
		writeJSON(w, map[string]any{"ok": false, "message": "ยังไม่พบ Python: ถ้าจะใช้ YOLO26 ให้ติดตั้ง Python และ ultralytics/opencv-python"})
		return
	}
	cmd := exec.Command(py, "-c", "import ultralytics, cv2; print('ok')")
	if out, err := cmd.CombinedOutput(); err != nil {
		writeJSON(w, map[string]any{"ok": false, "python": py, "message": "พบ Python แต่ยังไม่พบ ultralytics/opencv-python", "detail": string(out), "install": "pip install ultralytics opencv-python"})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "python": py, "message": "พร้อมใช้ YOLO26/OpenCV"})
}

func handleVehicleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer file.Close()
	os.MkdirAll(filepath.Join(cfg.PrinterPath, "vehicle_tmp"), 0755)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	imgPath := filepath.Join(cfg.PrinterPath, "vehicle_tmp", "detect_"+strconv.FormatInt(time.Now().UnixNano(), 10)+ext)
	if err := saveUploadedFile(file, imgPath); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	py := findPython()
	if py == "" {
		writeJSON(w, map[string]any{"ok": false, "error": "ไม่พบ Python", "install": "ติดตั้ง Python แล้วรัน pip install ultralytics opencv-python"})
		return
	}
	script := filepath.Join(appDir(), "yolo26_vehicle_detect.py")
	if !fileExists(script) {
		writeJSON(w, map[string]any{"ok": false, "error": "ไม่พบไฟล์ yolo26_vehicle_detect.py", "script": script})
		return
	}
	cmd := exec.Command(py, script, imgPath)
	cmd.Dir = appDir()
	out, err := cmd.CombinedOutput()
	raw := string(out)
	jsonText := extractJSONObject(raw)
	if err != nil {
		if json.Valid([]byte(jsonText)) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write([]byte(jsonText))
			return
		}
		writeJSON(w, map[string]any{"ok": false, "error": err.Error(), "raw": raw, "install": "pip install ultralytics opencv-python"})
		return
	}
	if !json.Valid([]byte(jsonText)) {
		writeJSON(w, map[string]any{"ok": false, "error": "Python ไม่ได้คืนค่า JSON ที่ถูกต้อง", "raw": raw})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(jsonText))
}


func extractJSONObject(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end >= start {
		return raw[start : end+1]
	}
	return raw
}

func findPython() string {
	for _, name := range []string{"python", "py", "python3"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func appDir() string {
	 ถ้าไฟล์ Python อยู่ในโฟลเดอร์ปัจจุบัน ให้ใช้โฟลเดอร์ปัจจุบันก่อน
	wd, _ := os.Getwd()
	if fileExists(filepath.Join(wd, "yolo26_vehicle_detect.py")) {
		return wd
	}
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		if fileExists(filepath.Join(exeDir, "yolo26_vehicle_detect.py")) {
			return exeDir
		}
	}
	return wd
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	kind := r.FormValue("kind")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer file.Close()
	name := header.Filename
	if kind == "template" {
		name = cfg.TemplateName
	} else if kind == "capture" {
		name = cfg.CaptureName
	}
	safe := filepath.Base(name)
	dst := filepath.Join(cfg.PrinterPath, safe)
	if err := saveUploadedFile(file, dst); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "saved": safe})
}

func saveUploadedFile(file multipart.File, dst string) error {
	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return err
}

func handleImage(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "template" {
		name = cfg.TemplateName
	}
	if name == "capture" {
		name = cfg.CaptureName
	}
	if name == "" {
		http.Error(w, "missing name", 400)
		return
	}
	path := filepath.Join(cfg.PrinterPath, filepath.Base(name))
	b, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "file not found: "+path, 404)
		return
	}
	lower := strings.ToLower(path)
	w.Header().Set("Cache-Control", "no-store")
	switch {
	case strings.HasSuffix(lower, ".tif") || strings.HasSuffix(lower, ".tiff"):
		w.Header().Set("Content-Type", "image/tiff")
	case strings.HasSuffix(lower, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Write(b)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func atoi(s string, def int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

var _ = atoi

const indexHTML = `<!doctype html>
<html lang="th">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>InspectionAgent V17</title>
<script src="https://cdn.jsdelivr.net/npm/tesseract.js@5/dist/tesseract.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/utif@3.1.0/UTIF.min.js"></script>
<style>
:root{--bg:#f5f7fb;--card:#fff;--ink:#182033;--muted:#667085;--line:#d9e0ea;--blue:#2563eb;--green:#16a34a;--red:#dc2626;--amber:#d97706;}
*{box-sizing:border-box} body{margin:0;background:var(--bg);font-family:Tahoma,Arial,sans-serif;color:var(--ink)}
header{padding:18px 24px;background:linear-gradient(135deg,#1d4ed8,#0f766e);color:white;box-shadow:0 4px 16px #0002} header h1{margin:0;font-size:24px} header p{margin:6px 0 0;opacity:.92}
.wrap{padding:18px;display:grid;grid-template-columns:360px 1fr;gap:16px}.card{background:var(--card);border:1px solid var(--line);border-radius:18px;padding:16px;box-shadow:0 8px 24px #10204012}.card h2{margin:0 0 12px;font-size:18px}.row{display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin:8px 0}.row label{font-size:13px;color:var(--muted);width:120px}.input,select{border:1px solid var(--line);border-radius:10px;padding:9px 10px;font-size:14px;min-width:120px;background:white}input[type=file]{width:100%}.btn{border:0;border-radius:12px;padding:10px 12px;cursor:pointer;font-weight:700;background:#e2e8f0;color:#111827}.btn:hover{filter:brightness(.97)}.primary{background:var(--blue);color:white}.success{background:var(--green);color:white}.danger{background:var(--red);color:white}.warn{background:#f59e0b;color:white}.ghost{background:#f8fafc;border:1px solid var(--line)}
.status{font-size:13px;padding:8px 10px;border-radius:10px;background:#f1f5f9;color:#334155;margin-top:8px}.ok{color:var(--green);font-weight:700}.bad{color:var(--red);font-weight:700}.small{font-size:12px;color:var(--muted)}
.viewer{overflow:auto;max-height:73vh;border:1px dashed #b8c2d0;border-radius:16px;background:#eef2f7;position:relative;padding:12px}.stage{position:relative;display:inline-block;transform-origin:top left}.stage canvas{display:block;max-width:none;border-radius:8px;box-shadow:0 6px 18px #0002;background:white}.box{position:absolute;border:2px solid #ff2d2d;background:#ff000018;cursor:move}.box .tag{position:absolute;left:0;top:-23px;background:#ef4444;color:white;font-size:12px;padding:2px 6px;border-radius:6px;white-space:nowrap}.box.active{border-color:#2563eb;background:#2563eb22}.box.active .tag{background:#2563eb}.handle{position:absolute;width:12px;height:12px;right:-7px;bottom:-7px;background:#111;border:2px solid #fff;border-radius:50%;cursor:nwse-resize}.toolbar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:10px}.resultGrid{display:grid;grid-template-columns:1fr 110px;gap:6px;margin-top:10px}.resultGrid div{background:#f8fafc;border:1px solid var(--line);border-radius:10px;padding:8px}.tabs{display:flex;gap:6px;margin-bottom:10px}.tab{padding:9px 12px;border-radius:12px;border:1px solid var(--line);cursor:pointer;background:#f8fafc}.tab.active{background:#dbeafe;border-color:#60a5fa;color:#1d4ed8;font-weight:700}.hide{display:none}.fileList{max-height:160px;overflow:auto;border:1px solid var(--line);border-radius:12px;background:#f8fafc}.fileItem{padding:7px 9px;border-bottom:1px solid #e5e7eb;cursor:pointer}.fileItem:hover{background:#e0f2fe}.fileItem b{display:block}.two{display:grid;grid-template-columns:1fr 1fr;gap:8px}.calcHero{background:linear-gradient(135deg,#eff6ff,#ecfdf5);border:1px solid #bfdbfe;border-radius:16px;padding:12px;margin-bottom:10px}.calcGrid{display:grid;grid-template-columns:1fr 1fr;gap:8px}.metric{border:1px solid var(--line);border-radius:14px;padding:10px;background:#f8fafc}.metric b{display:block;font-size:13px;color:#64748b;margin-bottom:4px}.metric span{font-size:22px;font-weight:800}.pass{background:#dcfce7!important;border-color:#22c55e!important;color:#14532d}.fail{background:#fee2e2!important;border-color:#ef4444!important;color:#7f1d1d}.warnbox{background:#fff7ed;border:1px solid #fed7aa;color:#9a3412;border-radius:12px;padding:9px;margin:8px 0}.brakeTable{width:100%;border-collapse:separate;border-spacing:0 7px}.brakeTable td{background:#f8fafc;border:1px solid #e2e8f0;padding:8px}.brakeTable td:first-child{border-radius:10px 0 0 10px;font-weight:700}.brakeTable td:last-child{border-radius:0 10px 10px 0;text-align:right;font-size:18px;font-weight:800}

.badge{display:inline-flex;align-items:center;gap:6px;border-radius:999px;padding:4px 10px;background:#e0f2fe;color:#075985;font-weight:800;font-size:12px}.titleIcon{font-size:22px;margin-right:8px}.input:focus,select:focus{outline:3px solid #bfdbfe;border-color:#3b82f6}.input.valuebox{font-weight:900;font-size:18px;text-align:right;background:linear-gradient(180deg,#ffffff,#f8fafc)}.unitPanel{background:linear-gradient(135deg,#fefce8,#eff6ff);border:1px solid #fde68a;border-radius:16px;padding:12px;margin:10px 0}.calcHero{box-shadow:inset 0 0 0 1px #ffffff99}.brakeTable td{box-shadow:0 4px 12px #0f172a0a}.brakeTable input{width:140px}.metric{position:relative;overflow:hidden;box-shadow:0 8px 18px #0f172a0c}.metric:before{content:"";position:absolute;left:0;top:0;width:6px;height:100%;background:#94a3b8}.metric.pass:before{background:#22c55e}.metric.fail:before{background:#ef4444}.metric b{font-size:14px}.metric span{letter-spacing:.2px}.formulaBox{background:#f8fafc;border:1px dashed #94a3b8;border-radius:14px;padding:10px;margin-top:10px;color:#334155;font-size:12px;line-height:1.65}.tab{font-weight:700}.tab.active{box-shadow:0 6px 14px #3b82f622}.card{backdrop-filter:blur(4px)}

.cameraPanel{display:none}.cameraPanel.active{display:block}.cameraBox{background:#0f172a;border-radius:18px;padding:12px;min-height:460px;color:white}.cameraStage{position:relative;display:inline-block;max-width:100%;background:#111;border-radius:14px;overflow:hidden}.cameraStage video,.cameraStage canvas{max-width:100%;border-radius:14px;display:block}.vehicleBox{position:absolute;border:5px solid #ef4444;border-radius:18px;box-shadow:0 0 0 9999px #00000018}.vehicleTag{position:absolute;left:10px;top:10px;color:#fff;font-weight:900;font-size:22px;text-shadow:0 2px 8px #000;background:#0008;border-radius:999px;padding:6px 14px}.cameraControls{display:grid;grid-template-columns:1fr 1fr;gap:8px;margin:10px 0}.camBadge{display:inline-block;padding:6px 12px;border-radius:999px;background:#e0f2fe;color:#075985;font-weight:900;margin:4px}.redBtn{background:#ef4444;color:#fff}.yellowBtn{background:#eab308;color:#111}.greenBtn{background:#22c55e;color:#052e16}
@media(max-width:960px){.wrap{grid-template-columns:1fr}.viewer{max-height:60vh}.cameraControls{grid-template-columns:1fr}}
</style>
</head>
<body>
<header><h1>InspectionAgent V17</h1><p>โครงสร้าง V12 เดิม + ปรับเฉพาะกล้อง: YOLO26/OpenCV/ป้ายทะเบียน</p></header>
<div class="wrap">
<aside class="card">
 <div class="tabs"><button class="tab active" onclick="showTab('ocr')">OCR เบรก</button><button class="tab" onclick="showTab('io')">CO/HC + เสียง</button><button class="tab" onclick="showTab('calc')">คำนวณเบรก</button><button class="tab" onclick="showTab('camera')">กล้องรถ</button><button class="tab" onclick="showTab('setting')">ตั้งค่า</button></div>
 <section id="tab-ocr">
  <h2>ไฟล์ผลเบรก</h2>
  <div class="status" id="fileStatus">กำลังตรวจไฟล์...</div>
  <div class="row"><button class="btn primary" onclick="loadImage('template')">เปิดแม่แบบ CalibrateValue</button><button class="btn success" onclick="loadImage('capture')">เปิด Capture</button></div>
  <div class="row"><button class="btn warn" onclick="readCaptureAuto()">อ่าน capture ด้วยกรอบแม่แบบ</button></div>
  <div class="small">หลักการ: ตั้งกรอบ 9 จุดบน CalibrateValue.tif แล้วกดบันทึก จากนั้นเมื่อมี capture.tif มา ให้เปิด capture แล้ว OCR ด้วยกรอบเดิม</div>
  <hr>
  <h2>เลือกไฟล์เอง</h2>
  <div class="row"><input type="file" id="uploadFile" accept=".tif,.tiff,.png,.jpg,.jpeg"><select id="uploadKind"><option value="capture">บันทึกเป็น capture.tif</option><option value="template">บันทึกเป็น CalibrateValue.tif</option><option value="keep">เก็บชื่อตามไฟล์</option></select></div>
  <button class="btn" onclick="uploadSelected()">อัปโหลดเข้า printerPath</button>
  <h2 style="margin-top:14px">ไฟล์ใน printerPath</h2>
  <div class="fileList" id="fileList"></div>
  <hr>
  <h2>กรอบ OCR 9 จุด</h2>
  <select id="boxSelect" class="input" onchange="selectBox(parseInt(this.value))"></select>
  <div class="row"><button class="btn" onclick="addOrResetBox()">สร้าง/รีเซ็ตกรอบนี้</button><button class="btn success" onclick="saveBoxes()">บันทึกกรอบ</button></div>
  <div class="row"><button class="btn primary" onclick="runOCRAll()">OCR 9 จุด</button><button class="btn ghost" onclick="clearResults()">ล้างผล</button></div>
  <div id="ocrStatus" class="status">พร้อมทำงาน</div>
  <div class="resultGrid" id="results"></div>
 </section>
 <section id="tab-io" class="hide">
  <h2>อ่านค่า CO/HC และเสียง</h2>
  <div class="row"><button class="btn primary" onclick="readCOHC()">อ่าน CO/HC</button><button class="btn primary" onclick="readSound()">อ่านเสียง</button></div>
  <div class="resultGrid"><div>CO</div><div id="coVal">-</div><div>HC</div><div id="hcVal">-</div><div>เสียง</div><div id="soundVal">-</div><div>RAW</div><div id="rawVal">-</div></div>
  <hr>
  <h2>ทดสอบ UDP Test Port</h2>
  <div class="row"><label>IP เครื่องมือ</label><input id="testUdpIp" class="input" value="192.168.1.145"></div>
  <div class="row"><label>Remote Port</label><input id="testUdpPort" class="input" value="7795"></div>
  <div class="row"><label>Local Port</label><input id="testLocalPort" class="input" value="2002"></div>
  <div class="row"><label>Command</label><input id="testCommand" class="input" value="e"></div>
  <button class="btn success" onclick="testUDP()">ส่ง UDP ทดสอบ</button>
  <div id="udpLog" class="status">ตัวอย่าง: IP 192.168.1.145 / Remote 7795 / Local 2002 / command e</div>
  <div class="small">ถ้าใช้ Packet Sender/Hercules ทดสอบ ให้ตั้งฝั่งเครื่องมือ/โปรแกรมจำลองให้รับ command แล้วตอบกลับมายัง local port 2002</div>
 </section>
 <section id="tab-calc" class="hide">
  <h2><span class="titleIcon">🧮</span>คำนวณผลเบรก</h2>
  <div class="calcHero">
   <span class="badge">✅ V11 Brake Calculator</span><br><br>
   <b>ใช้ค่าจาก OCR 9 จุด</b><br>
   <span class="small">ต้องเปิด/อ่าน capture.tif ก่อน แล้วกดปุ่มด้านล่างเพื่อดึงค่า OCR มาคำนวณ</span>
  </div>
  <div class="row"><button class="btn primary" onclick="fillCalcFromOCR()">รับค่าจาก OCR</button><button class="btn success" onclick="calculateBrake()">คำนวณ</button><button class="btn ghost" onclick="clearCalc()">ล้าง</button></div>
  <div id="calcStatus" class="status">ยังไม่ได้คำนวณ</div>
  <h2 style="margin-top:14px">🚗 แรงเบรกซ้าย/ขวา และน้ำหนัก</h2>
  <div class="unitPanel">
    <b>⚙️ หน่วยที่ใช้คำนวณ</b>
    <div class="row"><label>หน่วยแรงเบรก</label><select id="calcForceUnit" class="input"><option value="N">N</option><option value="daN">daN</option></select></div>
    <div class="row"><label>หน่วยน้ำหนัก</label><select id="calcWeightUnit" class="input"><option value="kgx10">kg×10 เช่น 10100 = 1010 kg</option><option value="kg">kg</option><option value="N">N</option></select></div>
    <div class="small">สูตรจะปรับตามหน่วยอัตโนมัติ เพื่อรองรับเครื่องที่ออกค่าเป็น N / daN / kg</div>
  </div>
  <table class="brakeTable">
   <tr><td>แรงเบรกหน้าซ้าย</td><td><input id="cFL" class="input valuebox" value=""></td></tr>
   <tr><td>แรงเบรกหน้าขวา</td><td><input id="cFR" class="input valuebox" value=""></td></tr>
   <tr><td>แรงเบรกหลังซ้าย</td><td><input id="cRL" class="input valuebox" value=""></td></tr>
   <tr><td>แรงเบรกหลังขวา</td><td><input id="cRR" class="input valuebox" value=""></td></tr>
   <tr><td>น้ำหนักหน้า</td><td><input id="cWF" class="input valuebox" value=""></td></tr>
   <tr><td>น้ำหนักหลัง</td><td><input id="cWR" class="input valuebox" value=""></td></tr>
   <tr><td>เบรกมือซ้าย</td><td><input id="cHL" class="input valuebox" value=""></td></tr>
   <tr><td>เบรกมือขวา</td><td><input id="cHR" class="input valuebox" value=""></td></tr>
  </table>
  <h2 style="margin-top:14px">📊 ผลคำนวณ</h2>
  <div class="calcGrid">
   <div id="mFrontDiff" class="metric"><b>ความแตกต่างล้อหน้า ≤ 25%</b><span>-</span></div>
   <div id="mRearDiff" class="metric"><b>ความแตกต่างล้อหลัง ≤ 25%</b><span>-</span></div>
   <div id="mServiceEff" class="metric"><b>ประสิทธิภาพเบรกหลัก 50–100%</b><span>-</span></div>
   <div id="mHandEff" class="metric"><b>ประสิทธิภาพเบรกมือ 20–100%</b><span>-</span></div>
  </div>
  <div id="overallResult" class="warnbox">รอผลคำนวณ</div>
  <div class="formulaBox" id="formulaExplain">สูตร: ความแตกต่าง = |ซ้าย-ขวา| / ค่ามากกว่า × 100<br>ประสิทธิภาพ = แรงเบรกรวม(N) / น้ำหนักรวม(N) × 100</div>
 </section>

 <section id="tab-camera" class="hide">
  <h2>📷 กล้อง Webcam / ภาพรถ</h2>
  <div class="status">แก้เฉพาะส่วนกล้องจาก V12: โหลดภาพ/เปิดกล้อง แล้วตรวจจับรถด้วย YOLO26 + OpenCV และหาตำแหน่งป้ายทะเบียนเบื้องต้น</div>
  <div class="row"><label>เลือกกล้อง</label><select id="cameraSelect" class="input" style="width:210px"></select></div>
  <div class="row"><button class="btn primary" onclick="refreshCameras()">ค้นหากล้อง</button><button class="btn success" onclick="openCamera()">เปิดกล้อง</button><button class="btn danger" onclick="stopCamera()">ปิดกล้อง</button></div>
  <div class="row"><button class="btn warn" onclick="captureCamera()">ถ่ายภาพจากกล้อง</button><button class="btn success" onclick="saveCameraImage()">Save ภาพ</button></div>
  <h2>โหลดภาพเพื่อทดสอบ</h2>
  <input type="file" id="vehicleFile" accept="image/*" onchange="loadVehicleImageFile()">
  <div class="row"><button class="btn primary" onclick="checkYOLO()">ตรวจสถานะ YOLO26</button><button class="btn success" onclick="detectVehicleYOLO()">ตรวจจับรถ + ป้ายทะเบียน</button></div>
  <div class="status" id="cameraStatus">ยังไม่ได้เปิดกล้อง</div>
  <h2>แก้ประเภทด้วยมือ ถ้าโมเดลอ่านผิด</h2>
  <div class="row"><button class="btn redBtn" onclick="drawVehicleBox('car')">รถยนต์ รย.1</button><button class="btn yellowBtn" onclick="drawVehicleBox('motorcycle')">จักรยานยนต์ รย.12</button></div>
  <div class="row"><button class="btn greenBtn" onclick="drawVehicleBox('van')">รถตู้ รย.2</button><button class="btn greenBtn" onclick="drawVehicleBox('pickup')">รถกระบะ รย.3</button></div>
  <div class="small">ติดตั้ง YOLO26: pip install ultralytics opencv-python แล้ววางโมเดล yolo26n.pt ไว้ข้างไฟล์ exe หรือปล่อยให้ Ultralytics โหลดโมเดลเอง</div>
 </section>
 <section id="tab-setting" class="hide">
  <h2>ตั้งค่า</h2>
  <div class="row"><label>printerPath</label><input id="printerPath" class="input" style="width:210px"></div>
  <h2>CO/HC</h2>
  <div class="row"><label>Mode</label><select id="cohcMode" class="input"><option>COM</option><option>UDP</option></select></div>
  <div class="row"><label>CO/HC COM</label><select id="cohcCom" class="input"></select></div>
  <div class="row"><label>UDP IP</label><input id="cohcUdpIp" class="input" value="192.168.1.145"></div>
  <div class="row"><label>Remote Port</label><input id="cohcUdpPort" class="input" value="7795"></div>
  <div class="row"><label>Local Port</label><input id="cohcLocalPort" class="input" value="2002"></div>
  <h2>เสียง</h2>
  <div class="row"><label>Mode</label><select id="soundMode" class="input"><option>COM</option><option>UDP</option></select></div>
  <div class="row"><label>Sound COM</label><select id="soundCom" class="input"></select></div>
  <div class="row"><label>UDP IP</label><input id="soundUdpIp" class="input" value="192.168.1.145"></div>
  <div class="row"><label>Remote Port</label><input id="soundUdpPort" class="input" value="7795"></div>
  <div class="row"><label>Local Port</label><input id="soundLocalPort" class="input" value="2002"></div>
  <div class="row"><label>UDP Timeout</label><input id="udpTimeoutMs" class="input" value="2000"><span>ms</span></div>
  <h2>หน่วยคำนวณเบรก</h2>
  <div class="row"><label>แรงเบรก</label><select id="brakeForceUnit" class="input"><option value="N">N</option><option value="daN">daN</option></select></div>
  <div class="row"><label>น้ำหนัก</label><select id="weightUnit" class="input"><option value="kgx10">kg×10 เช่น 10100</option><option value="kg">kg</option><option value="N">N</option></select></div>
  <h2>Serial รวม</h2>
  <div class="row"><label>Baudrate</label><select id="baudRate" class="input"><option>9600</option><option>19200</option><option>38400</option><option>57600</option><option>115200</option></select></div>
  <div class="row"><label>อ่านทุก</label><select id="readInterval" class="input"><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option></select><span>วินาที</span></div>
  <div class="row"><label>แม่แบบ</label><input id="templateName" class="input" value="CalibrateValue.tif"></div>
  <div class="row"><label>Capture</label><input id="captureName" class="input" value="capture.tif"></div>
  <button class="btn success" onclick="saveConfig()">บันทึกตั้งค่า</button>
  <div id="settingStatus" class="status">-</div>
 </section>
</aside>
<main class="card">
 <div id="cameraPanel" class="cameraPanel">
  <h2>📷 หน้ากล้อง / ภาพรถ</h2>
  <div class="cameraBox">
   <div class="cameraStage" id="cameraStage">
    <video id="cameraVideo" autoplay playsinline muted style="display:none"></video>
    <canvas id="vehicleCanvas" style="display:none;background:#111"></canvas>
    <div id="vehicleOverlay" class="vehicleBox" style="display:none"></div>
    <div id="vehicleTag" class="vehicleTag" style="display:none"></div>
   </div>
   <div style="margin-top:10px">
    <span class="camBadge">รถยนต์ รย.1 = กรอบแดง</span>
    <span class="camBadge">จักรยานยนต์ รย.12 = กรอบเหลือง</span>
    <span class="camBadge">รถตู้ รย.2 / กระบะ รย.3 = กรอบเขียว</span>
   </div>
  </div>
 </div>
 <div id="ocrPanel">
 <div class="toolbar"><button class="btn" onclick="zoomOut()">Zoom -</button><button class="btn" onclick="zoomReset()">100%</button><button class="btn" onclick="zoomIn()">Zoom +</button><span class="status" id="zoomLabel">100%</span><span class="status" id="currentFile">ยังไม่ได้เปิดภาพ</span></div>
 <div class="viewer" id="viewer"><div class="stage" id="stage"><canvas id="canvas" style="display:none;background:white;border-radius:8px;box-shadow:0 6px 18px #0002"></canvas></div></div>
</div>
</main>
</div>
<script>
const labels=['แรงเบรกหน้าซ้าย','แรงเบรกหน้าขวา','แรงเบรกหลังซ้าย','แรงเบรกหลังขวา','น้ำหนักหน้า','น้ำหนักหลัง','แรงเบรกมือซ้าย','แรงเบรกมือขวา','ศูนย์ล้อ'];
let cfg={}, files=[], currentName='', currentKind='', zoom=1, active=0;
let boxes = JSON.parse(localStorage.getItem('ocrBoxesV8')||'null') || labels.map((n,i)=>({x:0.08,y:0.08+i*0.055,w:0.16,h:0.045}));
let imgNatural={w:0,h:0};
let lastOcrValues={};
const canvas=document.getElementById('canvas'), ctxMain=canvas.getContext('2d'), stage=document.getElementById('stage');
function showTab(name){document.querySelectorAll('.tab').forEach(x=>x.classList.remove('active')); if(event&&event.target) event.target.classList.add('active'); ['ocr','io','calc','camera','setting'].forEach(t=>document.getElementById('tab-'+t).classList.toggle('hide',t!==name)); document.getElementById('cameraPanel').classList.toggle('active',name==='camera'); document.getElementById('ocrPanel').style.display=(name==='camera')?'none':'block'; if(name==='calc') fillCalcFromOCR(); if(name==='camera') refreshCameras();}
function initSelects(){let s=document.getElementById('boxSelect'); s.innerHTML=labels.map((l,i)=>'<option value="'+i+'">'+(i+1)+'. '+l+'</option>').join(''); ['cohcCom','soundCom'].forEach(id=>{let el=document.getElementById(id); el.innerHTML=''; for(let i=1;i<=30;i++) el.add(new Option('COM'+i,'COM'+i));}); renderResults();}
async function loadConfig(){cfg=await (await fetch('/api/config')).json(); ['printerPath','templateName','captureName'].forEach(k=>document.getElementById(k).value=cfg[k]||''); document.getElementById('cohcCom').value=cfg.cohcCom||'COM1'; document.getElementById('soundCom').value=cfg.soundCom||'COM2'; cohcMode.value=cfg.cohcMode||'COM'; soundMode.value=cfg.soundMode||'COM'; cohcUdpIp.value=cfg.cohcUdpIp||'192.168.1.145'; cohcUdpPort.value=cfg.cohcUdpPort||7795; cohcLocalPort.value=cfg.cohcLocalPort||2002; soundUdpIp.value=cfg.soundUdpIp||'192.168.1.145'; soundUdpPort.value=cfg.soundUdpPort||7795; soundLocalPort.value=cfg.soundLocalPort||2002; udpTimeoutMs.value=cfg.udpTimeoutMs||2000; testUdpIp.value=cfg.cohcUdpIp||'192.168.1.145'; testUdpPort.value=cfg.cohcUdpPort||7795; testLocalPort.value=cfg.cohcLocalPort||2002; document.getElementById('baudRate').value=cfg.baudRate||9600; document.getElementById('readInterval').value=cfg.readInterval||2; brakeForceUnit.value=cfg.brakeForceUnit||'N'; weightUnit.value=cfg.weightUnit||'kgx10'; calcForceUnit.value=cfg.brakeForceUnit||'N'; calcWeightUnit.value=cfg.weightUnit||'kgx10'; await refreshFiles();}
async function saveConfig(){let body={printerPath:printerPath.value,cohcCom:cohcCom.value,soundCom:soundCom.value,cohcMode:cohcMode.value,soundMode:soundMode.value,cohcUdpIp:cohcUdpIp.value,cohcUdpPort:+cohcUdpPort.value,cohcLocalPort:+cohcLocalPort.value,soundUdpIp:soundUdpIp.value,soundUdpPort:+soundUdpPort.value,soundLocalPort:+soundLocalPort.value,udpTimeoutMs:+udpTimeoutMs.value,baudRate:+baudRate.value,readInterval:+readInterval.value,templateName:templateName.value,captureName:captureName.value,brakeForceUnit:brakeForceUnit.value,weightUnit:weightUnit.value}; cfg=(await (await fetch('/api/config',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})).json()).config; settingStatus.innerHTML='บันทึกแล้ว'; await refreshFiles();}
async function refreshFiles(){let d=await (await fetch('/api/files?ts='+Date.now())).json(); files=d.files||[]; fileStatus.innerHTML='printerPath: <b>'+d.printerPath+'</b><br>แม่แบบ: '+(d.templateExists?'<span class=ok>พบ</span>':'<span class=bad>ไม่พบ</span>')+' / capture: '+(d.captureExists?'<span class=ok>พบ</span>':'<span class=bad>ไม่พบ</span>'); fileList.innerHTML=files.map(f=>'<div class="fileItem" onclick="loadImageByName(\''+f.name.replace(/'/g,'')+'\')"><b>'+f.name+'</b><span class=small>'+f.modified+' / '+f.size+' bytes</span></div>').join('') || '<div class=fileItem>ยังไม่มีไฟล์</div>';}
async function uploadSelected(){let f=uploadFile.files[0]; if(!f){alert('เลือกไฟล์ก่อน');return} let fd=new FormData(); fd.append('file',f); fd.append('kind',uploadKind.value); let r=await (await fetch('/api/upload',{method:'POST',body:fd})).json(); await refreshFiles(); await loadImageByName(r.saved);}
function imageUrl(name){return '/api/image?name='+encodeURIComponent(name)+'&ts='+Date.now()}
async function loadImage(kind){currentKind=kind; currentName=kind; currentFile.textContent=(kind==='template'?'แม่แบบ: ':'Capture: ')+(kind==='template'?cfg.templateName:cfg.captureName); await drawServerImage(kind);}
async function loadImageByName(name){currentKind='file'; currentName=name; currentFile.textContent='ไฟล์: '+name; await drawServerImage(name);}
async function drawServerImage(name){
 try{
  const url=imageUrl(name); const res=await fetch(url); if(!res.ok) throw new Error(await res.text());
  const buf=await res.arrayBuffer(); const lower=(name==='template'?cfg.templateName:name==='capture'?cfg.captureName:name).toLowerCase();
  if(lower.endsWith('.tif')||lower.endsWith('.tiff')){
    if(typeof UTIF==='undefined') throw new Error('ยังโหลด UTIF.js ไม่สำเร็จ ต้องมีอินเทอร์เน็ตครั้งแรก');
    const ifds=UTIF.decode(buf); UTIF.decodeImage(buf, ifds[0]); const rgba=UTIF.toRGBA8(ifds[0]);
    canvas.width=ifds[0].width; canvas.height=ifds[0].height; const imgData=new ImageData(new Uint8ClampedArray(rgba), canvas.width, canvas.height); ctxMain.putImageData(imgData,0,0);
  }else{
    const blob=new Blob([buf]); const obj=URL.createObjectURL(blob); const im=new Image(); await new Promise((ok,err)=>{im.onload=ok; im.onerror=err; im.src=obj}); canvas.width=im.naturalWidth; canvas.height=im.naturalHeight; ctxMain.drawImage(im,0,0); URL.revokeObjectURL(obj);
  }
  canvas.style.display='block'; imgNatural={w:canvas.width,h:canvas.height}; zoom=1; applyZoom(); renderBoxes();
 }catch(e){ alert('เปิดภาพไม่ได้: '+e.message); }
}
function applyZoom(){stage.style.transform='scale('+zoom+')'; zoomLabel.textContent=Math.round(zoom*100)+'%';}
function zoomIn(){zoom=Math.min(4,zoom+0.25); applyZoom();} function zoomOut(){zoom=Math.max(0.25,zoom-0.25); applyZoom();} function zoomReset(){zoom=1; applyZoom();}
function renderBoxes(){stage.querySelectorAll('.box').forEach(e=>e.remove()); if(!imgNatural.w)return; boxes.forEach((b,i)=>{let div=document.createElement('div'); div.className='box'+(i===active?' active':''); div.style.left=(b.x*imgNatural.w)+'px'; div.style.top=(b.y*imgNatural.h)+'px'; div.style.width=(b.w*imgNatural.w)+'px'; div.style.height=(b.h*imgNatural.h)+'px'; div.innerHTML='<span class=tag>'+(i+1)+'. '+labels[i]+'</span><span class=handle></span>'; div.onclick=(e)=>{e.stopPropagation();selectBox(i)}; makeDraggable(div,i); stage.appendChild(div);});}
function selectBox(i){active=i; boxSelect.value=i; renderBoxes();}
function addOrResetBox(){boxes[active]={x:0.1,y:0.1,w:0.18,h:0.05}; renderBoxes();}
function saveBoxes(){localStorage.setItem('ocrBoxesV8',JSON.stringify(boxes)); ocrStatus.innerHTML='บันทึกกรอบ OCR แล้ว ใช้กับ capture.tif ได้ทันที';}
function makeDraggable(el,i){let mode='', sx=0,sy=0,start; el.addEventListener('mousedown',e=>{mode=e.target.classList.contains('handle')?'resize':'move'; sx=e.clientX; sy=e.clientY; start={...boxes[i]}; e.preventDefault();}); window.addEventListener('mousemove',e=>{if(!mode)return; let dx=(e.clientX-sx)/(zoom*imgNatural.w), dy=(e.clientY-sy)/(zoom*imgNatural.h); if(mode==='move'){boxes[i].x=clamp(start.x+dx,0,1-start.w); boxes[i].y=clamp(start.y+dy,0,1-start.h);} else {boxes[i].w=clamp(start.w+dx,0.01,1-start.x); boxes[i].h=clamp(start.h+dy,0.01,1-start.y);} renderBoxes();}); window.addEventListener('mouseup',()=>mode='');}
function clamp(v,min,max){return Math.max(min,Math.min(max,v));}
function renderResults(vals={}){results.innerHTML=labels.map((l,i)=>'<div>'+(i+1)+'. '+l+'</div><div id="res'+i+'">'+(vals[i]||'-')+'</div>').join('');}
function clearResults(){renderResults();}
async function readCaptureAuto(){await loadImage('capture'); ocrStatus.innerHTML='โหลด capture แล้ว ตรวจตำแหน่งกรอบ จากนั้นกด OCR 9 จุด หรือรอภาพขึ้นแล้ว OCR'; setTimeout(runOCRAll,900);}
async function runOCRAll(){ if(!imgNatural.w){alert('เปิดภาพก่อน');return} saveBoxes(); ocrStatus.innerHTML='กำลัง OCR...'; const vals={}; for(let i=0;i<boxes.length;i++){selectBox(i); ocrStatus.innerHTML='กำลังอ่าน '+(i+1)+'/'+boxes.length+' : '+labels[i]; vals[i]=await ocrOne(boxes[i]); document.getElementById('res'+i).textContent=vals[i]||'-'; } lastOcrValues=vals; ocrStatus.innerHTML='อ่านครบแล้ว'; fillCalcFromOCR(false); }
async function ocrOne(b){let c=document.createElement('canvas'); let scale=3; c.width=Math.max(1,Math.round(b.w*imgNatural.w*scale)); c.height=Math.max(1,Math.round(b.h*imgNatural.h*scale)); let ctx=c.getContext('2d'); ctx.imageSmoothingEnabled=false; ctx.drawImage(canvas,b.x*imgNatural.w,b.y*imgNatural.h,b.w*imgNatural.w,b.h*imgNatural.h,0,0,c.width,c.height); let d=ctx.getImageData(0,0,c.width,c.height); for(let i=0;i<d.data.length;i+=4){let g=(d.data[i]*.299+d.data[i+1]*.587+d.data[i+2]*.114); let v=g>150?255:0; d.data[i]=d.data[i+1]=d.data[i+2]=v;} ctx.putImageData(d,0,0); let data=c.toDataURL('image/png'); let r=await Tesseract.recognize(data,'eng',{tessedit_char_whitelist:'0123456789.-'}); return cleanNum(r.data.text);}
function cleanNum(t){return (t||'').replace(/[^0-9.\-]/g,'').replace(/\.{2,}/g,'.').trim();}
async function readCOHC(){let d=await (await fetch('/api/cohc/read')).json(); coVal.textContent=d.co||'-'; hcVal.textContent=d.hc||'-'; rawVal.textContent=d.raw||d.error||'-';}
async function readSound(){let d=await (await fetch('/api/sound/read')).json(); soundVal.textContent=(d.sound||'-')+' '+(d.unit||''); rawVal.textContent=d.raw||d.error||'-';}
async function testUDP(){let body={ip:testUdpIp.value,port:+testUdpPort.value,localPort:+testLocalPort.value,command:testCommand.value,timeoutMs:+udpTimeoutMs.value||2000}; udpLog.innerHTML='กำลังส่ง UDP...'; let d=await (await fetch('/api/udp/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})).json(); udpLog.innerHTML=(d.ok?'<span class=ok>สำเร็จ</span>':'<span class=bad>ไม่สำเร็จ</span>')+'<br>SEND: '+body.command+' → '+body.ip+':'+body.port+' / local '+body.localPort+'<br>RECV: '+(d.raw||d.error||'-')+'<br>CO: '+(d.co||'-')+' HC: '+(d.hc||'-')+' SOUND: '+(d.sound||'-');}


let cameraStream=null, lastVehicleBlob=null, currentVehicleType='car', vehicleBaseDataUrl='';
async function refreshCameras(){
 try{
  if(!navigator.mediaDevices||!navigator.mediaDevices.enumerateDevices){cameraStatus.innerHTML='<span class="bad">Browser นี้ไม่รองรับกล้อง</span>';return;}
  const devices=await navigator.mediaDevices.enumerateDevices();
  const cams=devices.filter(d=>d.kind==='videoinput');
  cameraSelect.innerHTML='';
  cams.forEach((c,i)=>cameraSelect.add(new Option(c.label||('Camera '+(i+1)),c.deviceId)));
  cameraStatus.innerHTML='พบกล้อง '+cams.length+' ตัว';
 }catch(e){cameraStatus.innerHTML='<span class="bad">ค้นหากล้องไม่ได้: '+e.message+'</span>';}
}
async function openCamera(){
 try{
  stopCamera(false);
  const deviceId=cameraSelect.value;
  const opt={video: deviceId?{deviceId:{exact:deviceId},width:{ideal:1280},height:{ideal:720}}:{width:{ideal:1280},height:{ideal:720}}, audio:false};
  cameraStream=await navigator.mediaDevices.getUserMedia(opt);
  cameraVideo.srcObject=cameraStream;
  cameraVideo.style.display='block'; vehicleCanvas.style.display='none'; vehicleOverlay.style.display='none'; vehicleTag.style.display='none';
  cameraStatus.innerHTML='<span class="ok">เปิดกล้องแล้ว</span>';
 }catch(e){cameraStatus.innerHTML='<span class="bad">เปิดกล้องไม่ได้: '+e.message+'</span>';}
}
function stopCamera(msg=true){
 if(cameraStream){cameraStream.getTracks().forEach(t=>t.stop());cameraStream=null;}
 cameraVideo.style.display='none';
 if(msg) cameraStatus.innerHTML='ปิดกล้องแล้ว';
}
function captureCamera(){
 if(!cameraVideo.videoWidth){cameraStatus.innerHTML='<span class="bad">ยังไม่มีภาพจากกล้อง</span>';return;}
 vehicleCanvas.width=cameraVideo.videoWidth; vehicleCanvas.height=cameraVideo.videoHeight;
 const c=vehicleCanvas.getContext('2d'); c.drawImage(cameraVideo,0,0,vehicleCanvas.width,vehicleCanvas.height);
 vehicleCanvas.style.display='block'; cameraVideo.style.display='none';
 vehicleBaseDataUrl=vehicleCanvas.toDataURL('image/jpeg',0.92);
 vehicleCanvas.toBlob(b=>{lastVehicleBlob=b;},'image/jpeg',0.92);
 cameraStatus.innerHTML='<span class="ok">ถ่ายภาพแล้ว กดตรวจจับรถ + ป้ายทะเบียน หรือเลือกประเภทด้วยมือได้</span>';
}
function loadVehicleImageFile(){
 const f=vehicleFile.files[0]; if(!f)return;
 const img=new Image(); img.onload=()=>{
  vehicleCanvas.width=img.naturalWidth; vehicleCanvas.height=img.naturalHeight;
  vehicleCanvas.getContext('2d').drawImage(img,0,0);
  vehicleCanvas.style.display='block'; cameraVideo.style.display='none';
  vehicleBaseDataUrl=vehicleCanvas.toDataURL('image/jpeg',0.92);
  vehicleCanvas.toBlob(b=>{lastVehicleBlob=b;},'image/jpeg',0.92);
  cameraStatus.innerHTML='<span class="ok">โหลดภาพทดสอบแล้ว กดตรวจจับรถ + ป้ายทะเบียน</span>';
 };
 img.src=URL.createObjectURL(f);
}
async function checkYOLO(){
 try{let d=await (await fetch('/api/vehicle/yolo-status')).json(); cameraStatus.innerHTML=(d.ok?'<span class="ok">':'<span class="bad">')+d.message+'</span><br><span class="small">'+(d.install||'')+'</span>';}
 catch(e){cameraStatus.innerHTML='<span class="bad">ตรวจสถานะ YOLO ไม่ได้: '+e.message+'</span>';}
}
async function detectVehicleYOLO(){
 if(!lastVehicleBlob){cameraStatus.innerHTML='<span class="bad">ยังไม่มีภาพ ให้โหลดไฟล์หรือถ่ายภาพจากกล้องก่อน</span>';return;}
 cameraStatus.innerHTML='กำลังตรวจจับด้วย YOLO26/OpenCV...';
 const fd=new FormData(); fd.append('file',lastVehicleBlob,'vehicle.jpg');
 try{
  const d=await (await fetch('/api/vehicle/detect',{method:'POST',body:fd})).json();
  if(!d.ok){cameraStatus.innerHTML='<span class="bad">ตรวจจับไม่ได้: '+(d.error||d.message||'-')+'</span><br><span class="small">'+(d.install||'')+'</span>'; return;}
  drawDetections(d.detections||[], d.plates||[]);
  const v=(d.detections||[]).filter(x=>x.kind==='vehicle')[0];
  cameraStatus.innerHTML='<span class="ok">ตรวจจับสำเร็จ</span> '+(v?('ประเภท: <b>'+v.thai+'</b> / confidence '+Math.round((v.conf||0)*100)+'%'):'ไม่พบรถ')+'<br>ป้ายทะเบียนที่พบ: '+((d.plates||[]).length);
 }catch(e){cameraStatus.innerHTML='<span class="bad">เรียก YOLO ไม่ได้: '+e.message+'</span>';}
}
function restoreVehicleBase(callback){
 const img=new Image(); img.onload=()=>{vehicleCanvas.width=img.naturalWidth; vehicleCanvas.height=img.naturalHeight; vehicleCanvas.getContext('2d').drawImage(img,0,0); if(callback)callback();};
 if(vehicleBaseDataUrl) img.src=vehicleBaseDataUrl; else if(callback) callback();
}
function drawDetections(dets, plates){
 restoreVehicleBase(()=>{
  const ctx=vehicleCanvas.getContext('2d'); ctx.lineWidth=Math.max(4, Math.round(vehicleCanvas.width/240)); ctx.font='bold '+Math.max(22,Math.round(vehicleCanvas.width/38))+'px Segoe UI';
  dets.forEach(det=>{
   const b=det.box; const color=det.color||'#ef4444'; ctx.strokeStyle=color; ctx.fillStyle=color;
   ctx.strokeRect(b.x1,b.y1,b.x2-b.x1,b.y2-b.y1);
   const label=(det.thai||det.label||'vehicle')+' '+Math.round((det.conf||0)*100)+'%';
   const tw=ctx.measureText(label).width+22; ctx.fillRect(b.x1,Math.max(0,b.y1-42),tw,40); ctx.fillStyle=(det.type==='motorcycle')?'#111':'#fff'; ctx.fillText(label,b.x1+10,Math.max(28,b.y1-12));
  });
  ctx.lineWidth=Math.max(3, Math.round(vehicleCanvas.width/350)); ctx.font='bold '+Math.max(16,Math.round(vehicleCanvas.width/55))+'px Segoe UI';
  plates.forEach(p=>{const b=p.box; ctx.strokeStyle='#38bdf8'; ctx.fillStyle='#38bdf8'; ctx.strokeRect(b.x1,b.y1,b.x2-b.x1,b.y2-b.y1); ctx.fillText('ป้ายทะเบียน',b.x1,Math.max(20,b.y1-6));});
  vehicleCanvas.toBlob(b=>{lastVehicleBlob=b;},'image/jpeg',0.92);
 });
}
function drawVehicleBox(type){
 currentVehicleType=type;
 let text='รถยนต์ รย.1', color='#ef4444';
 if(type==='motorcycle'){text='จักรยานยนต์ รย.12'; color='#eab308';}
 if(type==='van'){text='รถตู้ รย.2'; color='#22c55e';}
 if(type==='pickup'){text='รถกระบะ รย.3'; color='#22c55e';}
 const base=(vehicleCanvas.style.display!=='none')?vehicleCanvas:cameraVideo;
 const w=base.clientWidth||640, h=base.clientHeight||360;
 vehicleOverlay.style.display='block'; vehicleTag.style.display='block';
 vehicleOverlay.style.borderColor=color;
 vehicleOverlay.style.left=Math.round(w*0.12)+'px'; vehicleOverlay.style.top=Math.round(h*0.18)+'px';
 vehicleOverlay.style.width=Math.round(w*0.76)+'px'; vehicleOverlay.style.height=Math.round(h*0.62)+'px';
 vehicleTag.style.background=color; vehicleTag.style.color=(type==='motorcycle')?'#111':'#fff'; vehicleTag.innerHTML=text;
 cameraStatus.innerHTML='ตีกรอบด้วยมือ: <b>'+text+'</b>';
}
async function saveCameraImage(){
 if(!lastVehicleBlob){cameraStatus.innerHTML='<span class="bad">ยังไม่มีภาพให้ Save</span>';return;}
 const fd=new FormData(); const name='vehicle_capture_'+new Date().toISOString().replace(/[:.]/g,'-')+'.jpg';
 fd.append('file',lastVehicleBlob,name); fd.append('kind','vehicle');
 const r=await (await fetch('/api/upload',{method:'POST',body:fd})).json();
 cameraStatus.innerHTML='<span class="ok">บันทึกภาพแล้ว: '+r.saved+'</span>'; await refreshFiles();
}

function numVal(id){let v=parseFloat((document.getElementById(id).value||'').replace(/,/g,'')); return isNaN(v)?0:v;}
function setMetric(id,value,pass){let el=document.getElementById(id); el.querySelector('span').textContent=value; el.classList.remove('pass','fail'); el.classList.add(pass?'pass':'fail');}
function fillCalcFromOCR(showMsg=true){
 const map=['cFL','cFR','cRL','cRR','cWF','cWR','cHL','cHR'];
 let got=false;
 map.forEach((id,i)=>{let v=(lastOcrValues && lastOcrValues[i]) || (document.getElementById('res'+i)?.textContent||'').trim(); if(v && v!=='-'){document.getElementById(id).value=v; got=true;}});
 if(showMsg) calcStatus.innerHTML=got?'ดึงค่าจาก OCR แล้ว กดคำนวณได้เลย':'ยังไม่มีค่า OCR กรุณาเปิด capture.tif แล้ว OCR 9 จุดก่อน';
}
function calculateBrake(){
 const fl=numVal('cFL'), fr=numVal('cFR'), rl=numVal('cRL'), rr=numVal('cRR'), wf=numVal('cWF'), wr=numVal('cWR'), hl=numVal('cHL'), hr=numVal('cHR');
 if(!fl||!fr||!rl||!rr||!wf||!wr){calcStatus.innerHTML='<span class="bad">ข้อมูลไม่ครบ ต้องมีแรงเบรกหน้า/หลัง และน้ำหนักหน้า/หลังก่อน</span>'; return;}
 const frontDiff=Math.abs(fl-fr)/Math.max(fl,fr)*100;
 const rearDiff=Math.abs(rl-rr)/Math.max(rl,rr)*100;
 const forceUnit=calcForceUnit.value||'N'; const weightUnit=calcWeightUnit.value||'kgx10';
 const forceMul = forceUnit==='daN'?10:1;
 let totalWeightN = 0;
 if(weightUnit==='kgx10') totalWeightN=((wf+wr)/10)*9.81;
 else if(weightUnit==='kg') totalWeightN=(wf+wr)*9.81;
 else totalWeightN=(wf+wr);
 const service=((fl+fr+rl+rr)*forceMul)/totalWeightN*100;
 const hand=(hl&&hr)?((hl+hr)*forceMul)/totalWeightN*100:0;
 formulaExplain.innerHTML='หน่วยแรงเบรก: <b>'+forceUnit+'</b> / หน่วยน้ำหนัก: <b>'+weightUnit+'</b><br>น้ำหนักรวม(N) = '+totalWeightN.toFixed(2)+'<br>แรงเบรกหลัก(N) = '+((fl+fr+rl+rr)*forceMul).toFixed(2)+' / แรงเบรกมือ(N) = '+((hl+hr)*forceMul).toFixed(2);
 const pFront=frontDiff<=25, pRear=rearDiff<=25, pService=service>=50&&service<=100, pHand=hand>=20&&hand<=100;
 setMetric('mFrontDiff',frontDiff.toFixed(2)+'%',pFront);
 setMetric('mRearDiff',rearDiff.toFixed(2)+'%',pRear);
 setMetric('mServiceEff',service.toFixed(2)+'%',pService);
 setMetric('mHandEff',hand?hand.toFixed(2)+'%':'-',pHand);
 const all=pFront&&pRear&&pService&&pHand;
 overallResult.className=all?'warnbox pass':'warnbox fail';
 overallResult.innerHTML=all?'ผ่านเกณฑ์ทั้งหมด':'มีบางรายการไม่ผ่านเกณฑ์ กรุณาตรวจสอบช่องสีแดง';
 calcStatus.innerHTML='คำนวณจากค่าแรงเบรกซ้าย/ขวาและน้ำหนักเรียบร้อย';
}
function clearCalc(){['cFL','cFR','cRL','cRR','cWF','cWR','cHL','cHR'].forEach(id=>document.getElementById(id).value=''); ['mFrontDiff','mRearDiff','mServiceEff','mHandEff'].forEach(id=>{let el=document.getElementById(id);el.querySelector('span').textContent='-';el.classList.remove('pass','fail');}); overallResult.className='warnbox'; overallResult.innerHTML='รอผลคำนวณ'; calcStatus.innerHTML='ล้างค่าแล้ว';}

initSelects(); loadConfig(); setInterval(refreshFiles,5000);
</script>
</body></html>`
