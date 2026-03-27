<script>
	import { template } from '$lib/state.svelte.js';
</script>

<div class="header-bar">
	<div class="header-left">
		<button class="bw header-type-btn">
			<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="3" y="3" width="18" height="18" rx="2"/>
				<path d="M3 9h18M9 3v18"/>
			</svg>
			Certificate
		</button>
		<div class="header-sep"></div>

		<select class="field-select" style="width:60px" bind:value={template.pageSize}>
			<option value="A4">A4</option>
			<option value="A3">A3</option>
			<option value="Letter">Letter</option>
		</select>

		<div class="btn-row">
			<button
				class="bw icon-sq"
				class:active={template.orientation === 'landscape'}
				onclick={() => template.orientation = 'landscape'}
				title="Landscape"
			>
				<svg width="13" height="10" viewBox="0 0 16 12"><rect x="0.5" y="0.5" width="15" height="11" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.2"/></svg>
			</button>
			<button
				class="bw icon-sq"
				class:active={template.orientation === 'portrait'}
				onclick={() => template.orientation = 'portrait'}
				title="Portrait"
			>
				<svg width="10" height="13" viewBox="0 0 12 16"><rect x="0.5" y="0.5" width="11" height="15" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.2"/></svg>
			</button>
		</div>
	</div>
	<div class="header-right">
		<span style="color:#666;font-size:10px">Certificate Generator</span>
		<a href="https://github.com/umarbektokyo/certificate-generator" target="_blank" rel="noopener noreferrer" class="bw icon-sq" title="GitHub" style="margin-left:8px;color:#888;display:inline-flex;align-items:center;justify-content:center;text-decoration:none;">
			<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
		</a>
	</div>
</div>
