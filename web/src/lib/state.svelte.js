/** @typedef {{id: string, key: string, x: number, y: number, fontSize: number, fontFamily: string, color: string, align: string, valign: string, bold: boolean, italic: boolean}} Field */
/** @typedef {{[key: string]: string}} Recipient */

let nextId = 1;
function uid() { return 'f' + nextId++; }

// --- Template ---
export let template = $state({
	name: 'Untitled',
	orientation: 'landscape', // landscape | portrait
	pageSize: 'A4',
	bgColor: '#ffffff',
	bgImage: /** @type {string|null} */ (null),
	bgFileName: /** @type {string|null} */ (null),
	bgFit: 'cover',   // cover | contain | stretch | original
	bgX: 50,           // horizontal position 0-100 (only for cover/contain)
	bgY: 50,           // vertical position 0-100
	bgScale: 100,      // additional scale percentage (100 = default)
	bgPdfData: /** @type {string|null} */ (null), // raw PDF base64 for server-side import
	bgPdfW: /** @type {number|null} */ (null),   // original PDF width in mm
	bgPdfH: /** @type {number|null} */ (null),   // original PDF height in mm
});

// --- Fields ---
export let fields = $state(/** @type {Field[]} */ ([
	{ id: uid(), key: 'name', x: 50, y: 45, fontSize: 32, fontFamily: 'serif', color: '#1a1a1a', align: 'center', valign: 'middle', bold: true, italic: false},
	{ id: uid(), key: 'course', x: 50, y: 58, fontSize: 16, fontFamily: 'sans-serif', color: '#444444', align: 'center', valign: 'middle', bold: false, italic: false},
	{ id: uid(), key: 'date', x: 50, y: 72, fontSize: 12, fontFamily: 'sans-serif', color: '#666666', align: 'center', valign: 'middle', bold: false, italic: false},
]));

export let selectedFieldId = $state({ value: /** @type {string|null} */ (null) });

export function addField(key = 'field') {
	const f = { id: uid(), key, x: 50, y: 50, fontSize: 14, fontFamily: 'sans-serif', color: '#333333', align: 'center', valign: 'middle', bold: false, italic: false};
	fields.push(f);
	for (const r of recipients) {
		if (!(key in r)) r[key] = '';
	}
	selectedFieldId.value = f.id;
	return f;
}

export function removeField(id) {
	const field = fields.find(f => f.id === id);
	if (field) {
		const key = field.key;
		fields.splice(fields.indexOf(field), 1);
		// Remove key from all recipients if no other field uses the same key
		if (!fields.some(f => f.key === key)) {
			for (const r of recipients) delete r[key];
		}
	}
	if (selectedFieldId.value === id) selectedFieldId.value = null;
}

export function getSelectedField() {
	return fields.find(f => f.id === selectedFieldId.value) ?? null;
}

// --- Recipients ---
export let recipients = $state(/** @type {Recipient[]} */ ([
	{ name: 'Your Name', course: 'Qualification', date: '2026-03-26' },
	{ name: 'Umarbek B', course: 'Or Unqualification', date: '2026-03-26' },
]));

export let previewIndex = $state({ value: 0 });

export function currentRecipient() {
	return recipients[previewIndex.value] ?? {};
}

// --- Custom fonts ---
/** @type {{name: string, data: string, css: string}[]} */
export let customFonts = $state([]);

export function addCustomFont(name, base64Data) {
	// Create CSS @font-face for preview
	const css = `@font-face { font-family: '${name}'; src: url(data:font/ttf;base64,${base64Data}) format('truetype'); }`;
	const style = document.createElement('style');
	style.textContent = css;
	document.head.appendChild(style);
	customFonts.push({ name, data: base64Data, css });
}

// --- Generation state ---
export let generating = $state({ value: false });
export let genProgress = $state({ value: 0, total: 0 });

// --- Page dimensions in mm ---
const PAGE_SIZES = {
	A4: { w: 297, h: 210 },
	A3: { w: 420, h: 297 },
	Letter: { w: 279.4, h: 215.9 },
};

export function getPageDimensions() {
	// Use actual PDF dimensions when a PDF background is loaded
	if (template.bgPdfW && template.bgPdfH) {
		return { w: template.bgPdfW, h: template.bgPdfH };
	}
	const base = PAGE_SIZES[template.pageSize] ?? PAGE_SIZES.A4;
	if (template.orientation === 'portrait') return { w: base.h, h: base.w };
	return { w: base.w, h: base.h };
}
