<script>
	import BlenderPanel from './BlenderPanel.svelte';
	import { fields, selectedFieldId, addField, removeField, getSelectedField, template, customFonts, addCustomFont } from '$lib/state.svelte.js';

	let newKey = $state('');
	let sel = $derived(getSelectedField());
	let pdfLoading = $state(false);

	function handleFontUpload(e) {
		const file = e.target.files?.[0];
		if (!file) return;
		const name = file.name.replace(/\.(ttf|otf)$/i, '');
		const reader = new FileReader();
		reader.onload = () => {
			const base64 = /** @type {string} */ (reader.result).split(',')[1];
			addCustomFont(name, base64);
			if (sel) sel.fontFamily = name;
		};
		reader.readAsDataURL(file);
		e.target.value = '';
	}

	function handleAddField() {
		const key = newKey.trim() || 'field';
		addField(key);
		newKey = '';
	}

	function handleBgUpload(e) {
		const file = e.target.files?.[0];
		if (!file) return;
		if (file.type === 'application/pdf') {
			handlePdfImport(file);
			e.target.value = '';
			return;
		}
		template.bgPdfData = null;
		template.bgPdfW = null;
		template.bgPdfH = null;
		const reader = new FileReader();
		reader.onload = () => {
			template.bgImage = /** @type {string} */ (reader.result);
			template.bgFileName = file.name;
		};
		reader.readAsDataURL(file);
		e.target.value = '';
	}

	async function handlePdfImport(file) {
		pdfLoading = true;
		try {
			const arrayBuf = await file.arrayBuffer();
			// Store raw PDF bytes for server-side vector import
			const uint8 = new Uint8Array(arrayBuf);
			let binary = '';
			for (let i = 0; i < uint8.length; i++) binary += String.fromCharCode(uint8[i]);
			template.bgPdfData = btoa(binary);
			// Render preview for the canvas editor
			const pdfjsLib = await loadPdfJs();
			const pdf = await pdfjsLib.getDocument({ data: arrayBuf }).promise;
			const page = await pdf.getPage(1);
			const viewport = page.getViewport({ scale: 2 });
			const canvas = document.createElement('canvas');
			canvas.width = viewport.width;
			canvas.height = viewport.height;
			await page.render({ canvasContext: canvas.getContext('2d'), viewport }).promise;
			template.bgImage = canvas.toDataURL('image/png');
			template.bgFileName = file.name;
			const base = page.getViewport({ scale: 1 });
			// PDF.js viewport units are CSS points (1/72 inch). Convert to mm.
			template.bgPdfW = base.width * 25.4 / 72;
			template.bgPdfH = base.height * 25.4 / 72;
			template.orientation = base.width > base.height ? 'landscape' : 'portrait';
		} catch (err) {
			alert('Failed to load PDF: ' + (err.message || err));
		} finally {
			pdfLoading = false;
		}
	}

	let pdfJsPromise = null;
	function loadPdfJs() {
		if (pdfJsPromise) return pdfJsPromise;
		pdfJsPromise = import('https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.9.155/pdf.min.mjs').then(mod => {
			mod.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/4.9.155/pdf.worker.min.mjs';
			return mod;
		});
		return pdfJsPromise;
	}

	function clearBg() {
		template.bgImage = null;
		template.bgFileName = null;
		template.bgPdfData = null;
		template.bgPdfW = null;
		template.bgPdfH = null;
	}
</script>

<div style="height:100%;overflow-y:auto;">
	<BlenderPanel title="Background" icon='<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>' open={true}>
		<div class="field-group">
			<span class="field-label">Image / PDF</span>
			<div style="display:flex;gap:4px;align-items:center;">
				<label class="bw btn-upload" style="flex:1">
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/></svg>
					{#if pdfLoading}Loading...{:else}{template.bgFileName ?? 'Upload'}{/if}
					<input type="file" accept="image/*,.pdf,application/pdf" style="display:none" onchange={handleBgUpload} />
				</label>
				{#if template.bgImage}
					<button class="bw icon-sq" onclick={clearBg} title="Remove" style="color:#e85050">
						<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M18 6L6 18M6 6l12 12"/></svg>
					</button>
				{/if}
			</div>
		</div>
		<div class="field-group">
			<span class="field-label">Color</span>
			<div style="display:flex;gap:4px;align-items:center;">
				<input type="color" class="color-swatch" bind:value={template.bgColor} />
				<input type="text" class="field-text" style="width:70px;font-family:monospace;font-size:10px" bind:value={template.bgColor} />
			</div>
		</div>
		{#if template.bgImage}
			<div class="field-group">
				<span class="field-label">Fit</span>
				<div class="btn-row">
					{#each [['cover','Cover'],['contain','Contain'],['stretch','Stretch'],['original','Original']] as [v,l]}
						<button class="bw btn-sm" class:active={(template.bgFit??'cover')===v} onclick={()=>template.bgFit=v}>{l}</button>
					{/each}
				</div>
			</div>
			<div class="prop-grid">
				<span class="prop-label">X pos</span>
				<div class="blender-slider-wrap">
					<div class="blender-slider-track">
						<div class="blender-slider-fill" style="width:{template.bgX??50}%"></div>
						<span class="blender-slider-text">{template.bgX??50}%</span>
					</div>
					<input type="range" class="blender-slider-input" min="0" max="100" step="1" bind:value={template.bgX} />
				</div>
				<span class="prop-label">Y pos</span>
				<div class="blender-slider-wrap">
					<div class="blender-slider-track">
						<div class="blender-slider-fill" style="width:{template.bgY??50}%"></div>
						<span class="blender-slider-text">{template.bgY??50}%</span>
					</div>
					<input type="range" class="blender-slider-input" min="0" max="100" step="1" bind:value={template.bgY} />
				</div>
				<span class="prop-label">Scale</span>
				<div class="blender-slider-wrap">
					<div class="blender-slider-track">
						<div class="blender-slider-fill" style="width:{((template.bgScale??100)-25)/175*100}%"></div>
						<span class="blender-slider-text">{template.bgScale??100}%</span>
					</div>
					<input type="range" class="blender-slider-input" min="25" max="200" step="1" bind:value={template.bgScale} />
				</div>
			</div>
		{/if}
	</BlenderPanel>

	<BlenderPanel title="Fields" icon='<path d="M4 7V4h16v3M9 20h6M12 4v16"/>' open={true}>
		{#each fields as field (field.id)}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<div
				class="list-row"
				class:selected={selectedFieldId.value === field.id}
				onclick={() => selectedFieldId.value = field.id}
				style="gap:6px;"
			>
				<span style="color:#4b76c2;font-size:10px;font-weight:600;width:12px;text-align:center">T</span>
				<span style="flex:1;overflow:hidden;text-overflow:ellipsis">{field.key}</span>
				<span style="color:#666;font-size:10px;font-variant-numeric:tabular-nums">{field.fontSize}pt</span>
				<button
					class="bw icon-sq"
					style="width:18px;height:18px;opacity:0;font-size:9px"
					onclick={(e) => { e.stopPropagation(); removeField(field.id); }}
					title="Remove"
				>
					<svg width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M18 6L6 18M6 6l12 12"/></svg>
				</button>
			</div>
		{/each}

		<div style="display:flex;gap:3px;margin-top:2px;">
			<input
				type="text" class="field-text" style="flex:1"
				placeholder="Field name..."
				bind:value={newKey}
				onkeydown={(e) => e.key === 'Enter' && handleAddField()}
			/>
			<button class="bw btn-sm" style="flex:none;width:auto;padding:3px 8px" onclick={handleAddField}>+ Add</button>
		</div>
	</BlenderPanel>

	{#if sel}
		<BlenderPanel title="Properties" icon='<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/>' open={true}>
			<div class="prop-grid">
				<span class="prop-label">Key</span>
				<input type="text" class="field-text" bind:value={sel.key} />

				<span class="prop-label">X %</span>
				<input type="number" class="field-text" min="0" max="100" step="0.5" bind:value={sel.x} />

				<span class="prop-label">Y %</span>
				<input type="number" class="field-text" min="0" max="100" step="0.5" bind:value={sel.y} />

				<span class="prop-label">Size</span>
				<div class="blender-slider-wrap">
					<div class="blender-slider-track">
						<div class="blender-slider-fill" style="width:{((sel.fontSize - 6) / 90) * 100}%"></div>
						<span class="blender-slider-text">{sel.fontSize} pt</span>
					</div>
					<input type="range" class="blender-slider-input" min="6" max="96" step="1" bind:value={sel.fontSize} />
				</div>

				<span class="prop-label">Font</span>
				<div style="display:flex;flex-direction:column;gap:3px;">
					<select class="field-select" bind:value={sel.fontFamily}>
						<optgroup label="Built-in (PDF)">
							<option value="sans-serif">Helvetica</option>
							<option value="serif">Times New Roman</option>
							<option value="monospace">Courier</option>
						</optgroup>
						<optgroup label="System">
							<option value="Arial" style="font-family:Arial">Arial</option>
							<option value="Georgia" style="font-family:Georgia">Georgia</option>
							<option value="Verdana" style="font-family:Verdana">Verdana</option>
							<option value="Trebuchet MS" style="font-family:Trebuchet MS">Trebuchet MS</option>
							<option value="Palatino" style="font-family:Palatino">Palatino</option>
							<option value="Garamond" style="font-family:Garamond">Garamond</option>
							<option value="Impact" style="font-family:Impact">Impact</option>
						</optgroup>
						{#if customFonts.length > 0}
							<optgroup label="Custom">
								{#each customFonts as cf}
									<option value={cf.name} style="font-family:'{cf.name}'">{cf.name}</option>
								{/each}
							</optgroup>
						{/if}
					</select>
					<label class="bw btn-sm" style="cursor:pointer;justify-content:center;height:20px;font-size:10px;">
						<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/></svg>
						Upload .ttf font
						<input type="file" accept=".ttf,.otf,font/ttf,font/otf" style="display:none" onchange={handleFontUpload} />
					</label>
				</div>

				<span class="prop-label">Color</span>
				<div style="display:flex;gap:4px;align-items:center;">
					<input type="color" class="color-swatch" bind:value={sel.color} />
					<input type="text" class="field-text" style="font-family:monospace;font-size:10px" bind:value={sel.color} />
				</div>

				<span class="prop-label">H-Align</span>
				<div class="btn-row">
					<button class="bw btn-sm" class:active={sel.align==='left'} onclick={()=>sel.align='left'} title="Align left">
						<svg width="12" height="10" viewBox="0 0 16 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 1h14M1 4.5h9M1 8h12M1 11.5h7"/></svg>
					</button>
					<button class="bw btn-sm" class:active={sel.align==='center'} onclick={()=>sel.align='center'} title="Align center">
						<svg width="12" height="10" viewBox="0 0 16 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 1h14M3.5 4.5h9M2 8h12M4.5 11.5h7"/></svg>
					</button>
					<button class="bw btn-sm" class:active={sel.align==='right'} onclick={()=>sel.align='right'} title="Align right">
						<svg width="12" height="10" viewBox="0 0 16 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 1h14M6 4.5h9M3 8h12M8 11.5h7"/></svg>
					</button>
				</div>

				<span class="prop-label">V-Align</span>
				<div class="btn-row">
					<button class="bw btn-sm" class:active={(sel.valign??'middle')==='top'} onclick={()=>sel.valign='top'} title="Align top">
						<svg width="12" height="10" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 1h10"/><path d="M6 3.5v7" stroke-dasharray="1.5 1.5"/></svg>
					</button>
					<button class="bw btn-sm" class:active={(sel.valign??'middle')==='middle'} onclick={()=>sel.valign='middle'} title="Align middle">
						<svg width="12" height="10" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 6h10"/><path d="M6 1v3.5M6 7.5v3.5" stroke-dasharray="1.5 1.5"/></svg>
					</button>
					<button class="bw btn-sm" class:active={(sel.valign??'middle')==='bottom'} onclick={()=>sel.valign='bottom'} title="Align bottom">
						<svg width="12" height="10" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M1 11h10"/><path d="M6 1.5v7" stroke-dasharray="1.5 1.5"/></svg>
					</button>
				</div>

				<span class="prop-label">Style</span>
				<div class="btn-row">
					<button class="bw btn-sm" class:active={sel.bold} onclick={()=>sel.bold=!sel.bold} style="font-weight:700">B</button>
					<button class="bw btn-sm" class:active={sel.italic} onclick={()=>sel.italic=!sel.italic} style="font-style:italic">I</button>
				</div>
			</div>
		</BlenderPanel>
	{/if}
</div>

<style>
	.list-row:hover button { opacity: 1 !important; }
</style>
