package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-pdf/fpdf"
)

type Field struct {
	Key      string  `json:"key"`
	X        float64 `json:"x"`        // percentage 0-100
	Y        float64 `json:"y"`        // percentage 0-100
	FontSize float64 `json:"fontSize"` // pt
	Font     string  `json:"font"`     // sans-serif, serif, monospace, or custom font name
	Color    string  `json:"color"`    // hex like #1a1a1a
	Align    string  `json:"align"`    // left, center, right
	VAlign   string  `json:"valign"`   // top, middle, bottom
	Bold     bool    `json:"bold"`
	Italic   bool    `json:"italic"`
}

type CustomFont struct {
	Name string `json:"name"` // display name
	Data string `json:"data"` // base64-encoded TTF file
}

type GenerateRequest struct {
	Background  string              `json:"background"`
	BgColor     string              `json:"bgColor"`
	BgFit       string              `json:"bgFit"`   // cover, contain, stretch, original
	BgX         float64             `json:"bgX"`     // 0-100 horizontal position
	BgY         float64             `json:"bgY"`     // 0-100 vertical position
	BgScale     float64             `json:"bgScale"` // scale multiplier (1.0 = 100%)
	Width       float64             `json:"width"`
	Height      float64             `json:"height"`
	CustomFonts []CustomFont        `json:"customFonts"`
	Fields      []Field             `json:"fields"`
	Recipients  []map[string]string `json:"recipients"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Serve static frontend if the directory exists (Docker production mode)
	staticDir := "./static"
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file; fall back to index.html for SPA routing
			path := staticDir + r.URL.Path
			if _, err := os.Stat(path); os.IsNotExist(err) && !strings.Contains(r.URL.Path, ".") {
				http.ServeFile(w, r, staticDir+"/index.html")
				return
			}
			fs.ServeHTTP(w, r)
		})
		log.Println("Serving static frontend from", staticDir)
	}

	log.Println("Certificate server listening on :8181")
	log.Fatal(http.ListenAndServe(":8181", corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	// Allow large payloads (background images can be big)
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Generate: %dx%dmm, %d fields, %d recipients, bg=%d bytes",
		int(req.Width), int(req.Height), len(req.Fields), len(req.Recipients), len(req.Background))

	if len(req.Recipients) == 0 {
		http.Error(w, "No recipients", http.StatusBadRequest)
		return
	}
	if len(req.Fields) == 0 {
		http.Error(w, "No fields", http.StatusBadRequest)
		return
	}

	// Decode background image if provided
	var bgImageBytes []byte
	var bgImageType string
	if req.Background != "" {
		var err error
		bgImageBytes, bgImageType, err = decodeDataURL(req.Background)
		if err != nil {
			http.Error(w, "Invalid background image: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Generate ZIP of PDFs
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	loadedFonts := make(map[string]bool)

	for i, recipient := range req.Recipients {
		pdfBytes, err := generatePDF(req, recipient, bgImageBytes, bgImageType, loadedFonts)
		if err != nil {
			http.Error(w, fmt.Sprintf("PDF generation failed for recipient %d: %v", i+1, err), http.StatusInternalServerError)
			return
		}

		// Use first field value as filename, fallback to index
		name := fmt.Sprintf("certificate_%d", i+1)
		for _, f := range req.Fields {
			if v, ok := recipient[f.Key]; ok && v != "" {
				name = sanitizeFilename(v)
				break
			}
		}

		fw, err := zipWriter.Create(name + ".pdf")
		if err != nil {
			http.Error(w, "ZIP error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fw.Write(pdfBytes)
	}

	zipWriter.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=certificates.zip")
	w.Write(zipBuf.Bytes())
}

func generatePDF(req GenerateRequest, recipient map[string]string, bgImageBytes []byte, bgImageType string, loadedFonts map[string]bool) ([]byte, error) {
	// Determine orientation from requested dimensions
	orient := "L"
	if req.Height > req.Width {
		orient = "P"
	}

	// fpdf Size defines the base page size in portrait; the OrientationStr swaps for landscape.
	// So always pass the smaller dimension as Wd and larger as Ht (portrait base).
	sizeW, sizeH := req.Width, req.Height
	if sizeW > sizeH {
		sizeW, sizeH = sizeH, sizeW
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: orient,
		UnitStr:        "mm",
		Size:           fpdf.SizeType{Wd: sizeW, Ht: sizeH},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	// Register custom fonts
	for _, cf := range req.CustomFonts {
		if loadedFonts[cf.Name] {
			continue
		}
		fontBytes, err := base64.StdEncoding.DecodeString(cf.Data)
		if err != nil {
			continue
		}
		pdf.AddUTF8FontFromBytes(cf.Name, "", fontBytes)
		pdf.AddUTF8FontFromBytes(cf.Name, "B", fontBytes)
		pdf.AddUTF8FontFromBytes(cf.Name, "I", fontBytes)
		pdf.AddUTF8FontFromBytes(cf.Name, "BI", fontBytes)
		loadedFonts[cf.Name] = true
	}

	pdf.AddPage()

	// The actual page dimensions after fpdf applies orientation
	pageW, pageH := pdf.GetPageSize()

	// Background color
	bgR, bgG, bgB := hexToRGB(req.BgColor)
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.Rect(0, 0, pageW, pageH, "F")

	// Background image
	if bgImageBytes != nil {
		registerAndPlaceImage(pdf, bgImageBytes, bgImageType, pageW, pageH, req.BgFit, req.BgX, req.BgY, req.BgScale)
	}

	// Render fields
	for _, field := range req.Fields {
		text := recipient[field.Key]
		if text == "" {
			continue
		}

		// Font style - map to PDF built-in fonts, custom fonts, or system font approximations
		fontFamily := "Helvetica"
		switch field.Font {
		case "serif", "Georgia", "Palatino", "Garamond":
			fontFamily = "Times"
		case "monospace":
			fontFamily = "Courier"
		case "sans-serif", "Arial", "Verdana", "Trebuchet MS", "Impact":
			fontFamily = "Helvetica"
		default:
			if loadedFonts[field.Font] {
				fontFamily = field.Font
			}
		}

		style := ""
		if field.Bold {
			style += "B"
		}
		if field.Italic {
			style += "I"
		}

		pdf.SetFont(fontFamily, style, field.FontSize)

		// Color
		cr, cg, cb := hexToRGB(field.Color)
		pdf.SetTextColor(cr, cg, cb)

		// Position (percentage to mm)
		x := (field.X / 100) * pageW
		y := (field.Y / 100) * pageH

		// Font size in mm: 1pt = 25.4/72 mm ≈ 0.3528mm
		fontMm := field.FontSize * 25.4 / 72.0

		// Alignment
		textWidth := pdf.GetStringWidth(text)

		var drawX float64
		switch field.Align {
		case "center":
			drawX = x - textWidth/2
		case "right":
			drawX = x - textWidth
		default:
			drawX = x
		}

		// pdf.Text Y is the baseline position.
		// Vertical alignment: top = baseline near top, middle = centered, bottom = baseline near bottom
		var drawY float64
		valign := field.VAlign
		if valign == "" {
			valign = "middle"
		}
		switch valign {
		case "top":
			drawY = y + fontMm*0.75 // baseline below top
		case "bottom":
			drawY = y // baseline at the point
		default: // middle
			drawY = y + fontMm*0.35
		}

		pdf.Text(drawX, drawY, text)

		if pdf.Err() {
			log.Printf("warning: pdf error after rendering field %q: %v", field.Key, pdf.Error())
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func registerAndPlaceImage(pdf *fpdf.Fpdf, imgBytes []byte, imgType string, pageW, pageH float64, fit string, posX, posY, sc float64) {
	reader := bytes.NewReader(imgBytes)
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		log.Printf("warning: could not decode image config: %v, placing at full page size", err)
		opt := fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
		pdf.RegisterImageOptionsReader("bg", opt, bytes.NewReader(imgBytes))
		pdf.ImageOptions("bg", 0, 0, pageW, pageH, false, opt, 0, "")
		return
	}

	// Defaults
	if fit == "" {
		fit = "cover"
	}
	if sc <= 0 {
		sc = 1.0
	}
	pxRatio := posX / 100.0 // 0..1
	pyRatio := posY / 100.0

	imgAspect := float64(cfg.Width) / float64(cfg.Height)
	pageAspect := pageW / pageH

	var drawW, drawH float64
	switch fit {
	case "stretch":
		drawW = pageW
		drawH = pageH
	case "original":
		// 1px = 25.4/96 mm at 96dpi; scale the raw pixel dims to mm
		drawW = float64(cfg.Width) * 25.4 / 96.0 * sc
		drawH = float64(cfg.Height) * 25.4 / 96.0 * sc
	case "contain":
		if imgAspect > pageAspect {
			drawW = pageW * sc
			drawH = drawW / imgAspect
		} else {
			drawH = pageH * sc
			drawW = drawH * imgAspect
		}
	default: // cover
		if imgAspect > pageAspect {
			drawH = pageH * sc
			drawW = drawH * imgAspect
		} else {
			drawW = pageW * sc
			drawH = drawW / imgAspect
		}
	}

	// Position: pxRatio/pyRatio controls where the image anchors
	drawX := (pageW - drawW) * pxRatio
	drawY := (pageH - drawH) * pyRatio

	opt := fpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
	pdf.RegisterImageOptionsReader("bg", opt, bytes.NewReader(imgBytes))
	pdf.ImageOptions("bg", drawX, drawY, drawW, drawH, false, opt, 0, "")
}

func decodeDataURL(dataURL string) ([]byte, string, error) {
	// data:image/png;base64,iVBOR...
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid data URL format")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		// Try URL-safe encoding
		data, err = base64.RawStdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("base64 decode failed: %w", err)
		}
	}

	// Determine image type from header
	imgType := "png"
	header := strings.ToLower(parts[0])
	if strings.Contains(header, "jpeg") || strings.Contains(header, "jpg") {
		imgType = "jpg"
	} else if strings.Contains(header, "gif") {
		imgType = "gif"
	}

	return data, imgType, nil
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string(hex[0]) + string(hex[0]) + string(hex[1]) + string(hex[1]) + string(hex[2]) + string(hex[2])
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_",
	)
	result := replacer.Replace(name)
	if len(result) > 100 {
		result = result[:100]
	}
	return result
}
