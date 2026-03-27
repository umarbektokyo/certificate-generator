/**
 * Parse CSV text into array of objects.
 * First row is treated as headers.
 * Handles quoted fields with commas and newlines.
 * @param {string} text
 * @returns {{headers: string[], rows: {[key: string]: string}[]}}
 */
export function parseCSV(text) {
	const lines = [];
	let current = '';
	let inQuotes = false;

	for (let i = 0; i < text.length; i++) {
		const ch = text[i];
		if (inQuotes) {
			if (ch === '"') {
				if (text[i + 1] === '"') {
					current += '"';
					i++;
				} else {
					inQuotes = false;
				}
			} else {
				current += ch;
			}
		} else {
			if (ch === '"') {
				inQuotes = true;
			} else if (ch === ',') {
				lines.push(current);
				current = '';
			} else if (ch === '\n' || (ch === '\r' && text[i + 1] === '\n')) {
				lines.push(current);
				current = '';
				if (ch === '\r') i++;
				lines.push(null); // row separator
			} else if (ch === '\r') {
				lines.push(current);
				current = '';
				lines.push(null);
			} else {
				current += ch;
			}
		}
	}
	if (current || lines.length > 0) lines.push(current);

	// Split into rows
	const rawRows = [];
	let row = [];
	for (const cell of lines) {
		if (cell === null) {
			if (row.length > 0) rawRows.push(row);
			row = [];
		} else {
			row.push(cell.trim());
		}
	}
	if (row.length > 0) rawRows.push(row);

	if (rawRows.length === 0) return { headers: [], rows: [] };

	const headers = rawRows[0].map(h => h.toLowerCase().trim());
	const rows = [];
	for (let i = 1; i < rawRows.length; i++) {
		const obj = {};
		for (let j = 0; j < headers.length; j++) {
			obj[headers[j]] = rawRows[i][j] ?? '';
		}
		// skip empty rows
		if (Object.values(obj).some(v => v !== '')) rows.push(obj);
	}

	return { headers, rows };
}
