<script>
	import { fields, recipients, template, getPageDimensions, generating, genProgress } from '$lib/state.svelte.js';
	let dims = $derived(getPageDimensions());
</script>

<div class="status-bar">
	<div class="status-left">
		<span>{fields.length} fields</span>
		<span class="status-sep">|</span>
		<span>{recipients.length} recipients</span>
		<span class="status-sep">|</span>
		<span>{template.pageSize} {template.orientation}</span>
		<span class="status-sep">|</span>
		<span>{dims.w}&times;{dims.h}mm</span>
	</div>
	<div class="status-right">
		{#if generating.value}
			<span style="color:#4b76c2">Generating {genProgress.value}/{genProgress.total}...</span>
		{:else}
			<span>Ready</span>
		{/if}
	</div>
</div>
