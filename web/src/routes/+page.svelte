<script>
	import Header from '$lib/components/Header.svelte';
	import StatusBar from '$lib/components/StatusBar.svelte';
	import FieldPanel from '$lib/components/FieldPanel.svelte';
	import Canvas from '$lib/components/Canvas.svelte';
	import DataPanel from '$lib/components/DataPanel.svelte';

	let leftW = $state(280);
	let rightW = $state(260);
	let resizing = $state(/** @type {null|'left'|'right'} */ (null));

	function startResize(side, e) {
		e.preventDefault();
		resizing = side;
		const startX = e.clientX;
		const startW = side === 'left' ? leftW : rightW;
		function onMove(e) {
			const dx = e.clientX - startX;
			const w = side === 'left' ? startW + dx : startW - dx;
			if (side === 'left') leftW = Math.max(200, Math.min(420, w));
			else rightW = Math.max(200, Math.min(420, w));
		}
		function onUp() {
			resizing = null;
			window.removeEventListener('mousemove', onMove);
			window.removeEventListener('mouseup', onUp);
		}
		window.addEventListener('mousemove', onMove);
		window.addEventListener('mouseup', onUp);
	}
</script>

<svelte:window onbeforeunload={(e) => { e.preventDefault(); }} />

<div class="mobile-warning">
	<div class="mobile-warning-card">
		<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#4b76c2" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
			<rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
			<line x1="8" y1="21" x2="16" y2="21"/>
			<line x1="12" y1="17" x2="12" y2="21"/>
		</svg>
		<h2>Desktop Required</h2>
		<p>The certificate editor needs a larger screen to work properly. Please open this page on a desktop or laptop computer, or resize your browser window wider.</p>
		<span class="mobile-warning-min">Minimum width: 900px</span>
	</div>
</div>

<div class="blender-app" class:cursor-col-resize={resizing}>
	<Header />
	<div class="blender-main">
		<div class="n-panel" style="width:{leftW}px">
			<FieldPanel />
		</div>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="region-handle" onmousedown={(e) => startResize('left', e)}></div>
		<Canvas />
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="region-handle" onmousedown={(e) => startResize('right', e)}></div>
		<div class="n-panel" style="width:{rightW}px">
			<DataPanel />
		</div>
	</div>
	<StatusBar />
</div>

<style>
	.cursor-col-resize, .cursor-col-resize * { cursor: col-resize !important; }
</style>
