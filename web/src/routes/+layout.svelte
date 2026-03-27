<script>
	import '../app.css';
	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap" rel="stylesheet" />

	<meta name="description" content="Design certificate templates with drag-and-drop, import recipients from CSV, and generate bulk PDFs instantly." />
	<meta name="keywords" content="certificate generator, bulk certificates, PDF generator, drag and drop, CSV import, template designer, award certificates, diploma generator" />
	<meta name="author" content="umarbektokyo" />
	<meta name="robots" content="index, follow" />
	<meta name="theme-color" content="#232323" />

	<meta property="og:title" content="Certificate Generator — Bulk PDF Certificates from CSV" />
	<meta property="og:description" content="Design certificate templates with drag-and-drop, import recipients from CSV, and generate bulk PDFs instantly." />
	<meta property="og:type" content="website" />
	<meta property="og:url" content="https://certificate-generator.umarbek.dev" />
	<meta property="og:image" content="https://certificate-generator.umarbek.dev/og.png" />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />
	<meta property="og:image:alt" content="Certificate Generator — Bulk PDF Certificates from CSV" />
	<meta property="og:site_name" content="Certificate Generator" />

	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="Certificate Generator — Bulk PDF Certificates from CSV" />
	<meta name="twitter:description" content="Design templates, import CSV, generate bulk PDF certificates." />
	<meta name="twitter:image" content="https://certificate-generator.umarbek.dev/og.png" />
	<meta name="twitter:image:alt" content="Certificate Generator — Bulk PDF Certificates from CSV" />
</svelte:head>

{@render children()}
