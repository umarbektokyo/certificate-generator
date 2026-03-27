# Certificate Generator

Bulk certificate generator with a visual drag-and-drop designer. Design once, generate hundreds of personalized PDF certificates from a CSV.

Built with SvelteKit + TailwindCSS (frontend) and Go (backend). Blender-inspired dark UI.

## Features

- **Visual designer** — drag text fields onto a certificate canvas, set position / font / size / color / alignment
- **Background support** — upload any image (PNG, JPEG) or use a solid color
- **CSV import** — bulk-load recipient data; new columns auto-create fields
- **Manual entry** — add / edit / remove recipients inline
- **Live preview** — navigate through recipients to preview each certificate
- **Bulk PDF generation** — Go backend generates all certificates as a ZIP of PDFs
- **Client-side PNG fallback** — works offline if the backend isn't running
- **Resizable panels** — Blender-style drag-to-resize region handles
- **Page presets** — A4, A3, Letter in landscape or portrait

## Architecture

```
certificate/
├── web/                    SvelteKit frontend (TailwindCSS v4)
│   └── src/
│       ├── lib/
│       │   ├── components/ UI components (Canvas, FieldPanel, DataPanel, ...)
│       │   ├── csv.js      Zero-dependency CSV parser
│       │   └── state.svelte.js  Svelte 5 runes state management
│       └── routes/         SvelteKit pages
├── server/                 Go backend
│   ├── main.go             HTTP server + PDF generation (go-pdf/fpdf)
│   └── main_test.go        Unit & integration tests
├── Dockerfile              Multi-stage build (Go + Node → single image)
└── docker-compose.yml      One-command deployment
```

## Quick Start

### Docker (recommended)

```bash
docker compose up --build
```

Open [http://localhost:8181](http://localhost:8181).

### Manual

**Prerequisites:** Node.js 20+, Go 1.22+

```bash
# Terminal 1: frontend dev server
cd web
npm install
npm run dev          # → http://localhost:5173

# Terminal 2: backend
cd server
go run .             # → http://localhost:8181
```

The Vite dev server proxies `/api` requests to the Go backend.

## Usage

1. **Set up the template** — upload a certificate background image (or pick a background color). Choose page size and orientation.
2. **Add fields** — each field maps to a column in your CSV (e.g. `name`, `course`, `date`). Drag fields on the canvas to position them. Adjust font, size, color, and alignment in the Properties panel.
3. **Import data** — click "Import CSV" and upload a file. The first row must be headers matching your field keys. Or add recipients manually.
4. **Preview** — navigate through recipients with the arrow buttons to verify placement.
5. **Generate** — click "Generate N Certificates". The backend produces a ZIP of PDFs. If the backend is offline, PNGs are generated client-side.

### CSV Format

```csv
name,course,date
Jane Doe,Web Development,2026-03-26
John Smith,Data Science,2026-03-26
```

Headers are case-insensitive and must match field keys. New columns that don't match existing fields are auto-created.

## API

### `GET /api/health`

Returns `ok`. Use for Docker health checks.

### `POST /api/generate`

Generate certificates and return a ZIP file.

**Request body (JSON):**

| Field        | Type                | Description                           |
|-------------|---------------------|---------------------------------------|
| `background` | `string \| null`   | Base64 data URL of background image   |
| `bgColor`    | `string`           | Hex background color (e.g. `#ffffff`) |
| `width`      | `number`           | Page width in mm                      |
| `height`     | `number`           | Page height in mm                     |
| `fields`     | `Field[]`          | Text fields to render                 |
| `recipients` | `object[]`         | Array of `{key: value}` maps          |

**Field object:**

| Field      | Type      | Description                              |
|-----------|-----------|------------------------------------------|
| `key`      | `string`  | Matches a key in recipient objects       |
| `x`        | `number`  | Horizontal position (0–100%)             |
| `y`        | `number`  | Vertical position (0–100%)               |
| `fontSize` | `number`  | Font size in pt                          |
| `font`     | `string`  | `sans-serif`, `serif`, or `monospace`    |
| `color`    | `string`  | Hex color                                |
| `align`    | `string`  | `left`, `center`, or `right`             |
| `bold`     | `boolean` | Bold text                                |
| `italic`   | `boolean` | Italic text                              |

**Response:** `application/zip` containing one PDF per recipient, named after the first field value.

## Testing

```bash
# Go backend (14 tests)
cd server && go test -v ./...

# Frontend CSV parser (14 tests)
cd web && npm test
```

## License

MIT
