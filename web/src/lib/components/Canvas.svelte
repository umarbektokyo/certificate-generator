<script>
	import { fields, selectedFieldId, template, getPageDimensions, currentRecipient } from '$lib/state.svelte.js';

	let containerEl = $state(/** @type {HTMLElement|null} */ (null));
	let dims = $derived(getPageDimensions());
	let recipient = $derived(currentRecipient());

	let containerW = $state(800);
	let containerH = $state(600);

	$effect(() => {
		if (!containerEl) return;
		const ro = new ResizeObserver(entries => {
			const r = entries[0].contentRect;
			containerW = r.width;
			containerH = r.height;
		});
		ro.observe(containerEl);
		return () => ro.disconnect();
	});

	// scale = pixels per mm
	let scale = $derived(() => {
		const pad = 40;
		const availW = containerW - pad * 2;
		const availH = containerH - pad * 2;
		if (availW <= 0 || availH <= 0) return 1;
		return Math.min(availW / dims.w, availH / dims.h);
	});

	let displayW = $derived(dims.w * scale());
	let displayH = $derived(dims.h * scale());

	// pt to mm: 1pt = 25.4/72 mm = 0.3528mm
	const PT_TO_MM = 25.4 / 72;

	// Drag
	let dragging = $state(/** @type {string|null} */ (null));
	let dragStart = $state({ mx: 0, my: 0, fx: 0, fy: 0 });

	function handleMouseDown(e, field) {
		e.preventDefault();
		e.stopPropagation();
		selectedFieldId.value = field.id;
		dragging = field.id;
		dragStart = { mx: e.clientX, my: e.clientY, fx: field.x, fy: field.y };
		document.body.style.cursor = 'grabbing';
		document.body.style.userSelect = 'none';
		window.addEventListener('mousemove', handleMouseMove);
		window.addEventListener('mouseup', handleMouseUp);
	}

	function handleMouseMove(e) {
		if (!dragging) return;
		const field = fields.find(f => f.id === dragging);
		if (!field) return;
		const dx = ((e.clientX - dragStart.mx) / displayW) * 100;
		const dy = ((e.clientY - dragStart.my) / displayH) * 100;
		field.x = Math.max(0, Math.min(100, dragStart.fx + dx));
		field.y = Math.max(0, Math.min(100, dragStart.fy + dy));
	}

	function handleMouseUp() {
		dragging = null;
		document.body.style.cursor = '';
		document.body.style.userSelect = '';
		window.removeEventListener('mousemove', handleMouseMove);
		window.removeEventListener('mouseup', handleMouseUp);
	}

	// Click on a field to select it; click on empty canvas to deselect
	function handleFieldClick(e, field) {
		e.stopPropagation();
		selectedFieldId.value = field.id;
	}

	function handleCanvasClick() {
		selectedFieldId.value = null;
	}

	function getFontFamily(ff) {
		if (ff === 'serif') return 'Georgia, "Times New Roman", serif';
		if (ff === 'sans-serif') return 'Helvetica, Arial, sans-serif';
		if (ff === 'monospace') return '"Courier New", monospace';
		return `"${ff}", sans-serif`;
	}

	function getFieldText(field) {
		return recipient[field.key] || `{${field.key}}`;
	}

	function tx(align) {
		if (align === 'center') return '-50%';
		if (align === 'right') return '-100%';
		return '0';
	}

	function ty(valign) {
		if (valign === 'middle') return '-50%';
		if (valign === 'bottom') return '-100%';
		return '0';
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="preview-area" bind:this={containerEl} onclick={handleCanvasClick}>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="cert-surface"
		style="width:{displayW}px;height:{displayH}px;background:{template.bgColor};"
	>
		{#if template.bgImage}
			{@const fit = template.bgFit ?? 'cover'}
			{@const sc = (template.bgScale ?? 100) / 100}
			{@const posX = template.bgX ?? 50}
			{@const posY = template.bgY ?? 50}
			<img
				src={template.bgImage}
				alt=""
				class="cert-bg"
				draggable="false"
				style="
					object-fit:{fit === 'stretch' ? 'fill' : fit === 'original' ? 'none' : fit};
					object-position:{posX}% {posY}%;
					{sc !== 1 ? `transform:scale(${sc});transform-origin:${posX}% ${posY}%;` : ''}
				"
			/>
		{/if}

		{#each fields as field (field.id)}
			{@const isSel = selectedFieldId.value === field.id}
			{@const isDrag = dragging === field.id}
			{@const noVal = !recipient[field.key]}
			{@const fontSizePx = field.fontSize * PT_TO_MM * scale()}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class="cert-field"
				class:cert-field-sel={isSel}
				style="
					left:{field.x}%;
					top:{field.y}%;
					transform:translate({tx(field.align)},{ty(field.valign ?? 'middle')});
					font-size:{fontSizePx}px;
					font-family:{getFontFamily(field.fontFamily)};
					color:{noVal ? 'rgba(75,118,194,0.5)' : field.color};
					text-align:{field.align};
					font-weight:{field.bold ? '700' : '400'};
					font-style:{field.italic ? 'italic' : 'normal'};
					cursor:{isDrag ? 'grabbing' : 'grab'};
				"
				onmousedown={(e) => handleMouseDown(e, field)}
				onclick={(e) => handleFieldClick(e, field)}
			>{getFieldText(field)}</div>
		{/each}

		{#if !template.bgImage && fields.length === 0}
			<div class="cert-empty">Upload a background and add fields to begin</div>
		{/if}
	</div>
</div>

<style>
	.cert-surface {
		position: relative;
		box-shadow: 0 2px 24px rgba(0,0,0,0.5);
		overflow: hidden;
	}
	.cert-bg {
		position: absolute; inset: 0;
		width: 100%; height: 100%;
		object-fit: cover;
		pointer-events: none;
		user-select: none;
	}
	.cert-field {
		position: absolute;
		user-select: none;
		white-space: nowrap;
		padding: 1px 3px;
		border: 1.5px solid transparent;
		border-radius: 2px;
		transition: border-color 0.06s;
	}
	.cert-field:hover {
		border-color: rgba(75, 118, 194, 0.4);
	}
	.cert-field-sel {
		border-color: #4b76c2 !important;
		background: rgba(75, 118, 194, 0.06);
	}
	.cert-empty {
		position: absolute; inset: 0;
		display: flex; align-items: center; justify-content: center;
		color: #555; font-size: 12px;
		pointer-events: none;
	}
</style>
