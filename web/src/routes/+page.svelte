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
