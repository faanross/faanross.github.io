<script lang="ts">
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});

	const sections = [
		{
			title: "Module 1: DLLs and Basic Loading",
			lessons: [
				{ slug: "module01-theory01", title: "Introduction to DLLs (Theory 1.1)" },
				{ slug: "module01-theory02", title: "Introduction to Shellcode (Theory 1.2)" },
				{ slug: "module01-theory03", title: "Standard DLL Loading in Windows (Theory 1.3)" },
				{ slug: "module01-lab01", title: "Create a Basic DLL (Lab 1.1)" },
				{ slug: "module01-lab02", title: "Create a Basic Loader in Go (Lab 1.2)" }
			]
		},
		{
			title: "Module 2: PE Format for Loaders",
			lessons: [
				{ slug: "module02-theory01", title: "PE File Structure Essentials (Theory 2.1)" },
				{ slug: "module02-theory02", title: "Addressing in PE Files (Theory 2.2)" },
				{ slug: "module02-lab01", title: "PE Header Inspection with PE-Bear (Lab 2.1)" },
				{ slug: "module02-lab02", title: "PE Header Parser in Go (Lab 2.2)" }
			]
		},
		{
			title: "Module 3: Reflective DLL Loading Core Logic",
			lessons: [
				{ slug: "module03-theory01", title: "Intro to Reflective DLL Loading (Theory 3.1)" },
				{ slug: "module03-theory02", title: "Memory Allocation (Theory 3.2)" },
				{ slug: "module03-theory03", title: "Mapping the DLL Image (Theory 3.3)" },
				{ slug: "module03-lab01", title: "Manual DLL Mapping in Go (Lab 3.1)" }
			]
		},
		{
			title: "Module 4: Handling Relocations and Imports",
			lessons: [
				{ slug: "module04-theory01", title: "Base Relocations (Theory 4.1)" },
				{ slug: "module04-theory02", title: "IAT Resolution (Theory 4.2)" },
				{ slug: "module04-lab01", title: "Intentional Base Relocation (Lab 4.1)" },
				{ slug: "module04-lab02", title: "IAT Processing (Lab 4.2)" }
			]
		},
		{
			title: "Module 5: Execution and Exports",
			lessons: [
				{ slug: "module05-theory01", title: "The DLL Entry Point (Theory 5.1)" },
				{ slug: "module05-theory02", title: "Exported Functions (Theory 5.2)" },
				{ slug: "module05-lab01", title: "Call DllMain (Lab 5.1)" },
				{ slug: "module05-lab02", title: "Call Exported Function (Lab 5.2)" }
			]
		},
		{
			title: "Module 6: Basic Obfuscation - XOR",
			lessons: [
				{ slug: "module06-theory01", title: "Introduction to Obfuscation (Theory 6.1)" },
				{ slug: "module06-theory02", title: "Simple XOR (Theory 6.2)" },
				{ slug: "module06-lab01", title: "XOR Functions in Go (Lab 6.1)" },
				{ slug: "module06-lab02", title: "Obfuscated Loading (Lab 6.2)" }
			]
		},
		{
			title: "Module 7: Rolling XOR & Key Derivation",
			lessons: [
				{ slug: "module07-theory01", title: "Rolling XOR (Theory 7.1)" },
				{ slug: "module07-theory02", title: "Key Derivation Logic (Theory 7.2)" },
				{ slug: "module07-lab01", title: "Implementing Rolling XOR (Lab 7.1)" },
				{ slug: "module07-lab02", title: "Implementing Key Derivation (Lab 7.2)" }
			]
		},
		{
			title: "Module 8: Network Delivery & Client/Server",
			lessons: [
				{ slug: "module08-theory01", title: "Client + Server Communication (Theory 8.1)" },
				{ slug: "module08-theory02", title: "Communication Protocol Design (Theory 8.2)" },
				{ slug: "module08-theory03", title: "Environmental Keying + Client ID (Theory 8.3)" },
				{ slug: "module08-lab01", title: "Client + Server Logic (Lab 8.1)" },
				{ slug: "module08-lab02", title: "Implement Client ID and Key Derivation (Lab 8.2)" }
			]
		}
	];
</script>

<svelte:head>
	<title>Let's Build a Reflective Loader in Golang | Faan Rossouw</title>
	<meta name="description" content="Comprehensive course on building a reflective DLL loader in Golang, covering PE parsing, memory mapping, relocations, and obfuscation techniques." />
</svelte:head>

<section class="course-hero">
	<div class="container">
		{#if mounted}
			<a href="/courses" class="back-link" in:fly={{ y: -10, duration: 400, delay: 100 }}>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<polyline points="15 18 9 12 15 6"></polyline>
				</svg>
				Back to Courses
			</a>
			<span class="badge" in:fly={{ y: 20, duration: 600, delay: 200 }}>Faan Ross</span>
			<h1 in:fly={{ y: 20, duration: 600, delay: 300 }}>Let's Build a Reflective Loader in Golang</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Comprehensive course covering DLLs, PE format, reflective loading, relocations, imports, and obfuscation techniques.
			</p>
			<div class="meta" in:fly={{ y: 20, duration: 600, delay: 500 }}>
				<span class="date">Self-Paced</span>
				<span class="separator">|</span>
				<span class="duration">8 Modules</span>
			</div>
		{/if}
	</div>
</section>

<section class="course-overview">
	<div class="container">
		{#if mounted}
			<div class="overview-card glass-card" in:fly={{ y: 30, duration: 600, delay: 700 }}>
				<h2>Overview</h2>
				<p>
					This course walks through building a reflective DLL loader from the ground up. Starting with DLL fundamentals and shellcode basics,
					we progress through PE file structure, manual memory mapping, handling relocations and imports, and ultimately executing loaded code.
					The course also covers obfuscation techniques including XOR encryption with rolling keys and network-based payload delivery.
				</p>
			</div>
		{/if}
	</div>
</section>

<section class="course-toc">
	<div class="container">
		{#if mounted}
			{#each sections as section, sectionIndex}
				<div class="toc-section" in:fly={{ y: 30, duration: 600, delay: 800 + sectionIndex * 100 }}>
					<h3>{section.title}</h3>
					<ul>
						{#each section.lessons as lesson}
							<li>
								<a href="/courses/reflective/{lesson.slug}">{lesson.title}</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		{/if}
	</div>
</section>

<style>
	.course-hero {
		padding: 80px 0 40px;
		text-align: center;
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.6);
		text-decoration: none;
		margin-bottom: 24px;
		transition: color 0.3s ease;
	}

	.back-link:hover {
		color: var(--aion-purple);
	}

	.badge {
		display: inline-block;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		padding: 4px 12px;
		background: rgba(245, 230, 99, 0.2);
		border: 1px solid rgba(245, 230, 99, 0.4);
		border-radius: 12px;
		color: var(--aion-yellow);
		margin-bottom: 16px;
	}

	.course-hero h1 {
		font-size: clamp(28px, 4vw, 42px);
		font-weight: 700;
		margin-bottom: 16px;
		max-width: 800px;
		margin-left: auto;
		margin-right: auto;
	}

	.lead {
		font-size: clamp(16px, 2vw, 18px);
		color: rgba(255, 255, 255, 0.7);
		max-width: 600px;
		margin: 0 auto 20px;
	}

	.meta {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.5);
		margin-bottom: 24px;
	}

	.separator {
		margin: 0 12px;
	}

	.date {
		color: var(--aion-purple);
	}

	.course-overview {
		padding: 20px 0 40px;
	}

	.overview-card {
		max-width: 800px;
		margin: 0 auto;
		padding: 32px;
	}

	.overview-card h2 {
		font-size: 20px;
		margin-bottom: 12px;
	}

	.overview-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.7;
	}

	.course-toc {
		padding: 20px 0 80px;
		max-width: 800px;
		margin: 0 auto;
	}

	.toc-section {
		margin-bottom: 32px;
	}

	.toc-section h3 {
		font-size: 18px;
		font-weight: 600;
		color: var(--aion-purple);
		margin-bottom: 12px;
		padding-bottom: 8px;
		border-bottom: 1px solid rgba(189, 147, 249, 0.2);
	}

	.toc-section ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.toc-section li {
		margin-bottom: 8px;
	}

	.toc-section a {
		display: block;
		padding: 12px 16px;
		background: rgba(61, 61, 71, 0.3);
		border: 1px solid rgba(189, 147, 249, 0.1);
		border-radius: 8px;
		color: rgba(255, 255, 255, 0.8);
		text-decoration: none;
		font-size: 14px;
		transition: all 0.3s ease;
	}

	.toc-section a:hover {
		background: rgba(61, 61, 71, 0.5);
		border-color: rgba(189, 147, 249, 0.3);
		color: var(--white);
		transform: translateX(4px);
	}

	@media (max-width: 768px) {
		.course-hero {
			padding: 60px 0 30px;
		}

		.overview-card {
			padding: 24px;
		}

		.course-toc {
			padding: 20px 24px 60px;
		}
	}
</style>
