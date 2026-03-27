package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Unit tests for helper functions ---

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex        string
		r, g, b    int
	}{
		{"#ffffff", 255, 255, 255},
		{"#000000", 0, 0, 0},
		{"#ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"ffffff", 255, 255, 255},  // without #
		{"#fff", 255, 255, 255},    // shorthand
		{"#f00", 255, 0, 0},       // shorthand
		{"#1a1a1a", 26, 26, 26},
		{"", 0, 0, 0},             // empty
		{"#zzzzzz", 0, 0, 0},     // invalid
		{"#12", 0, 0, 0},          // too short
	}

	for _, tt := range tests {
		r, g, b := hexToRGB(tt.hex)
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("hexToRGB(%q) = (%d,%d,%d), want (%d,%d,%d)", tt.hex, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"John Doe", "John Doe"},
		{"file/name", "file_name"},
		{"a\\b:c*d?e\"f<g>h|i", "a_b_c_d_e_f_g_h_i"},
		{"normal.pdf", "normal.pdf"},
		{strings.Repeat("a", 150), strings.Repeat("a", 100)}, // truncation
		{"", ""},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDecodeDataURL(t *testing.T) {
	// Valid PNG data URL
	payload := base64.StdEncoding.EncodeToString([]byte("fakepng"))
	dataURL := "data:image/png;base64," + payload
	data, imgType, err := decodeDataURL(dataURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if imgType != "png" {
		t.Errorf("imgType = %q, want %q", imgType, "png")
	}
	if string(data) != "fakepng" {
		t.Errorf("data = %q, want %q", string(data), "fakepng")
	}

	// JPEG detection
	_, imgType, err = decodeDataURL("data:image/jpeg;base64," + payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if imgType != "jpg" {
		t.Errorf("imgType = %q, want %q", imgType, "jpg")
	}

	// GIF detection
	_, imgType, err = decodeDataURL("data:image/gif;base64," + payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if imgType != "gif" {
		t.Errorf("imgType = %q, want %q", imgType, "gif")
	}

	// Invalid format (no comma)
	_, _, err = decodeDataURL("nodataurl")
	if err == nil {
		t.Error("expected error for invalid data URL")
	}

	// Invalid base64
	_, _, err = decodeDataURL("data:image/png;base64,!!!invalid!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

// --- Integration tests for HTTP handlers ---

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func TestGenerateEndpoint_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/generate", nil)
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestGenerateEndpoint_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/generate", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGenerateEndpoint_NoRecipients(t *testing.T) {
	body := `{"fields":[{"key":"name"}],"recipients":[],"bgColor":"#fff","width":297,"height":210}`
	req := httptest.NewRequest("POST", "/api/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGenerateEndpoint_NoFields(t *testing.T) {
	body := `{"fields":[],"recipients":[{"name":"John"}],"bgColor":"#fff","width":297,"height":210}`
	req := httptest.NewRequest("POST", "/api/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGenerateEndpoint_Success(t *testing.T) {
	payload := GenerateRequest{
		BgColor: "#ffffff",
		Width:   297,
		Height:  210,
		Fields: []Field{
			{Key: "name", X: 50, Y: 50, FontSize: 24, Font: "sans-serif", Color: "#000000", Align: "center", Bold: true},
			{Key: "course", X: 50, Y: 65, FontSize: 14, Font: "serif", Color: "#444444", Align: "center", Italic: true},
		},
		Recipients: []map[string]string{
			{"name": "Alice", "course": "Go Programming"},
			{"name": "Bob", "course": "Web Development"},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/zip")
	}

	// Verify ZIP contents
	zipReader, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}

	if len(zipReader.File) != 2 {
		t.Errorf("zip contains %d files, want 2", len(zipReader.File))
	}

	// Files should be named after the first field value
	names := make(map[string]bool)
	for _, f := range zipReader.File {
		names[f.Name] = true
	}
	if !names["Alice.pdf"] {
		t.Errorf("expected Alice.pdf in zip, got %v", names)
	}
	if !names["Bob.pdf"] {
		t.Errorf("expected Bob.pdf in zip, got %v", names)
	}
}

func TestGenerateEndpoint_WithBackgroundImage(t *testing.T) {
	// Create a small valid PNG
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURL := "data:image/png;base64," + b64

	payload := GenerateRequest{
		Background: dataURL,
		BgColor:    "#ffffff",
		Width:      297,
		Height:     210,
		Fields:     []Field{{Key: "name", X: 50, Y: 50, FontSize: 20, Font: "sans-serif", Color: "#000", Align: "left"}},
		Recipients: []map[string]string{{"name": "Test User"}},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	// Verify the PDF is valid (starts with %PDF)
	zipReader, err := zip.NewReader(bytes.NewReader(rec.Body.Bytes()), int64(rec.Body.Len()))
	if err != nil {
		t.Fatalf("failed to read zip: %v", err)
	}

	f, err := zipReader.File[0].Open()
	if err != nil {
		t.Fatalf("failed to open zip entry: %v", err)
	}
	defer f.Close()

	pdfBytes, _ := io.ReadAll(f)
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Error("generated file does not start with %PDF header")
	}
}

func TestGenerateEndpoint_PortraitOrientation(t *testing.T) {
	payload := GenerateRequest{
		BgColor:    "#ffffff",
		Width:      210,
		Height:     297, // portrait: height > width
		Fields:     []Field{{Key: "name", X: 50, Y: 50, FontSize: 20, Font: "monospace", Color: "#000", Align: "right", Bold: true, Italic: true}},
		Recipients: []map[string]string{{"name": "Portrait Test"}},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/generate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleGenerate(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// OPTIONS preflight
	req := httptest.NewRequest("OPTIONS", "/api/generate", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("OPTIONS status = %d, want 200", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Errorf("CORS origin = %q, want %q", v, "*")
	}

	// Normal request passes through
	req = httptest.NewRequest("GET", "/test", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
	if v := rec.Header().Get("Access-Control-Allow-Methods"); v != "POST, GET, OPTIONS" {
		t.Errorf("CORS methods = %q", v)
	}
}

// --- Unit test for generatePDF ---

func TestGeneratePDF_AllFontFamilies(t *testing.T) {
	fonts := []string{"sans-serif", "serif", "monospace", "unknown"}
	for _, font := range fonts {
		req := GenerateRequest{BgColor: "#ffffff", Width: 297, Height: 210}
		recipient := map[string]string{"name": "Test"}
		field := Field{Key: "name", X: 50, Y: 50, FontSize: 16, Font: font, Color: "#000", Align: "center"}
		req.Fields = []Field{field}

		pdfBytes, err := generatePDF(req, recipient, nil, "", make(map[string]bool))
		if err != nil {
			t.Errorf("generatePDF with font %q failed: %v", font, err)
			continue
		}
		if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
			t.Errorf("font %q: output is not a valid PDF", font)
		}
	}
}

func TestGeneratePDF_EmptyFieldValue(t *testing.T) {
	req := GenerateRequest{BgColor: "#ffffff", Width: 297, Height: 210}
	req.Fields = []Field{{Key: "name", X: 50, Y: 50, FontSize: 16, Font: "sans-serif", Color: "#000", Align: "left"}}
	recipient := map[string]string{"name": ""} // empty value should be skipped

	pdfBytes, err := generatePDF(req, recipient, nil, "", make(map[string]bool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Error("output is not a valid PDF")
	}
}
