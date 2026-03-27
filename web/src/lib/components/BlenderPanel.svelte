<script>
	let { title, icon = '', open = true, children } = $props();
	let isOpen = $state(open);
</script>

<div class="bp">
	<button class="bp-header" onclick={() => isOpen = !isOpen}>
		<svg class="bp-disc" class:open={isOpen} width="10" height="10" viewBox="0 0 10 10">
			<path d="M3 1l4 4-4 4" fill="none" stroke="currentColor" stroke-width="1.5"/>
		</svg>
		{#if icon}
			<svg class="bp-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				{@html icon}
			</svg>
		{/if}
		{title}
	</button>
	{#if isOpen}
		<div class="bp-body">
			{@render children()}
		</div>
	{/if}
</div>
