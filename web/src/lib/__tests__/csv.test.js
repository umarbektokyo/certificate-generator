import { describe, it, expect } from 'vitest';
import { parseCSV } from '../csv.js';

describe('parseCSV', () => {
	it('parses basic CSV with headers', () => {
		const { headers, rows } = parseCSV('name,course\nAlice,Math\nBob,Science');
		expect(headers).toEqual(['name', 'course']);
		expect(rows).toEqual([
			{ name: 'Alice', course: 'Math' },
			{ name: 'Bob', course: 'Science' },
		]);
	});

	it('handles quoted fields with commas', () => {
		const { rows } = parseCSV('name,address\nAlice,"123 Main St, Apt 4"');
		expect(rows[0].address).toBe('123 Main St, Apt 4');
	});

	it('handles quoted fields with escaped quotes', () => {
		const { rows } = parseCSV('name,note\nAlice,"She said ""hello"""');
		expect(rows[0].note).toBe('She said "hello"');
	});

	it('handles CRLF line endings', () => {
		const { rows } = parseCSV('name\r\nAlice\r\nBob\r\n');
		expect(rows).toHaveLength(2);
		expect(rows[0].name).toBe('Alice');
		expect(rows[1].name).toBe('Bob');
	});

	it('handles CR-only line endings', () => {
		const { rows } = parseCSV('name\rAlice\rBob');
		expect(rows).toHaveLength(2);
	});

	it('trims whitespace from cells and headers', () => {
		const { headers, rows } = parseCSV(' name , course \n Alice , Math ');
		expect(headers).toEqual(['name', 'course']);
		expect(rows[0]).toEqual({ name: 'Alice', course: 'Math' });
	});

	it('lowercases headers', () => {
		const { headers } = parseCSV('Name,COURSE\nAlice,Math');
		expect(headers).toEqual(['name', 'course']);
	});

	it('skips empty rows', () => {
		const { rows } = parseCSV('name\nAlice\n\n\nBob');
		expect(rows).toHaveLength(2);
	});

	it('handles missing columns gracefully', () => {
		const { rows } = parseCSV('name,course\nAlice');
		expect(rows[0]).toEqual({ name: 'Alice', course: '' });
	});

	it('returns empty for empty input', () => {
		const { headers, rows } = parseCSV('');
		expect(headers).toEqual([]);
		expect(rows).toEqual([]);
	});

	it('returns empty rows for header-only input', () => {
		const { headers, rows } = parseCSV('name,course');
		expect(headers).toEqual(['name', 'course']);
		expect(rows).toEqual([]);
	});

	it('handles single column', () => {
		const { headers, rows } = parseCSV('name\nAlice\nBob');
		expect(headers).toEqual(['name']);
		expect(rows).toEqual([{ name: 'Alice' }, { name: 'Bob' }]);
	});

	it('handles quoted fields with newlines inside', () => {
		const { rows } = parseCSV('name,bio\nAlice,"Line 1\nLine 2"');
		expect(rows[0].bio).toBe('Line 1\nLine 2');
	});

	it('handles many columns', () => {
		const { headers, rows } = parseCSV('a,b,c,d,e\n1,2,3,4,5');
		expect(headers).toHaveLength(5);
		expect(Object.keys(rows[0])).toHaveLength(5);
	});
});
