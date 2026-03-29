<script>
	import BlenderPanel from './BlenderPanel.svelte';
	import JSZip from 'jszip';
	import { parseCSV } from '$lib/csv.js';
	import { fields, recipients, previewIndex, generating, genProgress, template, getPageDimensions, customFonts } from '$lib/state.svelte.js';

	let csvError = $state('');
	let quality = $state('pdf');

	function handleCSVUpload(e) {
		const file = e.target.files?.[0];
		if (!file) return;
		csvError = '';
		const reader = new FileReader();
		reader.onload = () => {
			try {
				const { headers, rows } = parseCSV(/** @type {string} */ (reader.result));
				if (rows.length === 0) { csvError = 'No data rows found'; return; }
				recipients.length = 0;
				recipients.push(...rows);
				previewIndex.value = 0;
				const existingKeys = new Set(fields.map(f => f.key));
				let yOff = 50;
				for (const h of headers) {
					if (!existingKeys.has(h)) {
						fields.push({
							id: 'f' + Date.now() + Math.random().toString(36).slice(2, 5),
							key: h, x: 50, y: yOff, fontSize: 14, fontFamily: 'sans-serif',
							color: '#333333', align: 'center', valign: 'middle', bold: false, italic: false,
						});
						yOff += 8;
					}
				}
			} catch { csvError = 'Failed to parse CSV'; }
		};
		reader.readAsText(file);
		e.target.value = '';
	}

	function addRecipient() {
		const obj = {};
		for (const f of fields) obj[f.key] = '';
		recipients.push(obj);
		previewIndex.value = recipients.length - 1;
	}

	function removeRecipient(idx) {
		recipients.splice(idx, 1);
		if (previewIndex.value >= recipients.length) previewIndex.value = Math.max(0, recipients.length - 1);
	}

	async function generateCertificates() {
		if (recipients.length === 0 || fields.length === 0) return;
		generating.value = true;
		genProgress.value = 0;
		genProgress.total = recipients.length;

		if (quality === 'pdf') {
			await generatePDF();
		} else {
			await clientGen();
		}
		generating.value = false;
	}

	async function generatePDF() {
		const dims = getPageDimensions();
		const usedCustom = new Set(fields.map(f => f.fontFamily).filter(f => !['sans-serif','serif','monospace'].includes(f)));
		const fontsPayload = customFonts.filter(cf => usedCustom.has(cf.name)).map(cf => ({ name: cf.name, data: cf.data }));
		const payload = {
			background: template.bgPdfData ? null : (template.bgImage || null),
			bgPdf: template.bgPdfData || null,
			bgColor: template.bgColor,
			bgFit: template.bgFit ?? 'cover',
			bgX: template.bgX ?? 50,
			bgY: template.bgY ?? 50,
			bgScale: (template.bgScale ?? 100) / 100,
			width: dims.w, height: dims.h,
			customFonts: fontsPayload,
			fields: fields.map(f => ({ key: f.key, x: f.x, y: f.y, fontSize: f.fontSize, font: f.fontFamily, color: f.color, align: f.align, valign: f.valign ?? 'middle', bold: f.bold, italic: f.italic })),
			recipients,
		};
		try {
			const res = await fetch('/api/generate', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
			if (!res.ok) throw new Error(await res.text());
			dl(await res.blob(), 'certificates.zip');
		} catch {
			alert('PDF server unavailable. Select a PNG quality option or start the Go server.');
		}
	}

	function drawBgImage(ctx, img, cw, ch) {
		const fit = template.bgFit ?? 'cover';
		const posX = (template.bgX ?? 50) / 100;
		const posY = (template.bgY ?? 50) / 100;
		const sc = (template.bgScale ?? 100) / 100;
		const ia = img.width / img.height;
		const ca = cw / ch;
		let dw, dh;
		if (fit === 'stretch') { dw = cw; dh = ch; }
		else if (fit === 'original') { dw = img.width * sc; dh = img.height * sc; }
		else if (fit === 'contain') {
			if (ia > ca) { dw = cw * sc; dh = dw / ia; }
			else { dh = ch * sc; dw = dh * ia; }
		} else { // cover
			if (ia > ca) { dh = ch * sc; dw = dh * ia; }
			else { dw = cw * sc; dh = dw / ia; }
		}
		const dx = (cw - dw) * posX;
		const dy = (ch - dh) * posY;
		ctx.drawImage(img, dx, dy, dw, dh);
	}

	function dl(blob, name) {
		const a = document.createElement('a');
		a.href = URL.createObjectURL(blob);
		a.download = name; a.click();
		URL.revokeObjectURL(a.href);
	}

	async function clientGen() {
		const dims = getPageDimensions();
		const scaleMap = { low: 3, medium: 6, high: 10, ultra: 14 };
		const s = scaleMap[quality] ?? 10;
		const cw = Math.round(dims.w * s), ch = Math.round(dims.h * s);
		const canvas = document.createElement('canvas');
		canvas.width = cw; canvas.height = ch;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		let bgImg = null;
		if (template.bgImage) {
			bgImg = await new Promise(r => { const i = new Image(); i.onload = () => r(i); i.onerror = () => r(null); i.src = template.bgImage; });
		}
		const zip = new JSZip();
		for (let i = 0; i < recipients.length; i++) {
			genProgress.value = i + 1;
			const r = recipients[i];
			ctx.fillStyle = template.bgColor; ctx.fillRect(0, 0, cw, ch);
			if (bgImg) { drawBgImage(ctx, bgImg, cw, ch); }
			for (const f of fields) { const v=r[f.key]; if(!v)continue; const x=(f.x/100)*cw,y=(f.y/100)*ch,sz=f.fontSize*s*0.3528; let fm='Inter,sans-serif'; if(f.fontFamily==='serif')fm='Georgia,serif'; if(f.fontFamily==='monospace')fm='Courier New,monospace'; ctx.font=`${f.italic?'italic ':''}${f.bold?'700':'400'} ${sz}px ${fm}`; ctx.fillStyle=f.color; ctx.textAlign=/** @type{CanvasTextAlign}*/(f.align); const vb=f.valign==='top'?'top':f.valign==='bottom'?'bottom':'middle'; ctx.textBaseline=/** @type{CanvasTextBaseline}*/(vb); ctx.fillText(v,x,y); }
			const blob = await new Promise(r => canvas.toBlob(r, 'image/png'));
			const name = (r[fields[0]?.key] || `certificate_${i + 1}`) + '.png';
			zip.file(name, blob);
		}
		const zipBlob = await zip.generateAsync({ type: 'blob' });
		dl(zipBlob, 'certificates.zip');
	}
</script>

<div style="height:100%;overflow-y:auto;">
	<BlenderPanel title="Data" icon='<path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/>' open={true}>
		<label class="bw btn-upload" style="width:100%">
			<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
			Import CSV
			<input type="file" accept=".csv,.tsv,text/csv" style="display:none" onchange={handleCSVUpload} />
		</label>
		{#if csvError}
			<span style="color:#e85050;font-size:10px">{csvError}</span>
		{/if}
		<span style="color:#555;font-size:10px;line-height:1.4">
			Headers = field keys. New columns auto-create fields.
		</span>
	</BlenderPanel>

	<BlenderPanel title="Recipients ({recipients.length})" icon='<path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>' open={true}>
		{#if recipients.length > 0}
			<div style="display:flex;align-items:center;gap:3px;">
				<button class="bw icon-sq" onclick={() => { if (previewIndex.value > 0) previewIndex.value-- }} disabled={previewIndex.value === 0} title="Previous">
					<svg width="8" height="10" viewBox="0 0 8 10" fill="currentColor"><path d="M7 1L1 5l6 4z"/></svg>
				</button>
				<span style="flex:1;text-align:center;font-variant-numeric:tabular-nums;color:#ccc">
					{previewIndex.value + 1} / {recipients.length}
				</span>
				<button class="bw icon-sq" onclick={() => { if (previewIndex.value < recipients.length - 1) previewIndex.value++ }} disabled={previewIndex.value >= recipients.length - 1} title="Next">
					<svg width="8" height="10" viewBox="0 0 8 10" fill="currentColor"><path d="M1 1l6 4-6 4z"/></svg>
				</button>
			</div>

			<div style="max-height:200px;overflow-y:auto;border-radius:4px;">
				{#each recipients as r, i}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<div
						class="list-row"
						class:selected={previewIndex.value === i}
						onclick={() => previewIndex.value = i}
					>
						<span style="color:#555;width:18px;text-align:right;flex-shrink:0;font-variant-numeric:tabular-nums;font-size:10px">{i+1}</span>
						<span style="flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{Object.values(r)[0] || '(empty)'}</span>
						<button
							class="bw icon-sq"
							style="width:16px;height:16px;opacity:0;font-size:8px"
							onclick={(e) => { e.stopPropagation(); removeRecipient(i); }}
							title="Remove"
						>
							<svg width="8" height="8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M18 6L6 18M6 6l12 12"/></svg>
						</button>
					</div>
				{/each}
			</div>
		{/if}

		<button class="bw btn-sm" style="width:100%" onclick={addRecipient}>+ Add Recipient</button>
	</BlenderPanel>

	{#if recipients.length > 0 && recipients[previewIndex.value]}
		<BlenderPanel title="Edit" icon='<path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>' open={true}>
			<div class="prop-grid">
				{#each Object.keys(recipients[previewIndex.value]) as key}
					<span class="prop-label">{key}</span>
					<input type="text" class="field-text" bind:value={recipients[previewIndex.value][key]} />
				{/each}
			</div>
		</BlenderPanel>
	{/if}

	<BlenderPanel title="Export" icon='<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>' open={true}>
		<div style="display:flex;align-items:center;justify-content:space-between;gap:4px;">
			<span class="prop-label" style="flex-shrink:0">Quality</span>
			<span style="color:#888;font-size:10px;flex-shrink:0">{{ pdf:'Vector PDF', low:'~890×630', medium:'~1780×1260', high:'~2970×2100', ultra:'~4160×2940' }[quality]}</span>
		</div>
		<div class="btn-row" style="display:flex;gap:2px;">
			{#each [['pdf','PDF'],['low','Low'],['medium','Med'],['high','High'],['ultra','Ultra']] as [v,l]}
				<button class="bw btn-sm" style="flex:1" class:active={quality===v} onclick={()=>quality=v}>{l}</button>
			{/each}
		</div>
		<button
			class="bw btn-action"
			disabled={generating.value || recipients.length === 0 || fields.length === 0}
			onclick={generateCertificates}
		>
			<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
			{#if generating.value}
				Generating {genProgress.value}/{genProgress.total}...
			{:else}
				Download {recipients.length} Certificate{recipients.length !== 1 ? 's' : ''}
			{/if}
		</button>
		<span style="color:#555;font-size:10px">
			ZIP of PNGs. Higher quality = larger file size.
		</span>
	</BlenderPanel>
</div>

<style>
	.list-row:hover button { opacity: 1 !important; }
</style>
