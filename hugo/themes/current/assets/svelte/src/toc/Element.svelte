<svelte:options customElement={{
	tag: 'dpb587-toc',
	shadow: 'none',
	props: {
		forRef: { type: 'String', attribute: 'for' },
		buttonClass: { type: 'String', attribute: 'button-class' }
	}
}} />

<script>
import { fade, fly } from 'svelte/transition';

let buttonClass = "";
let buttonActive = false;

function handleButtonClick(e) {
	buttonActive = !buttonActive;

	if (!buttonActive) {
		e.currentTarget.blur();
	}
}

function scrollToTop(e) {
	window.scroll({ top: 0, left: 0 });

	try {
		history.pushState("", document.title, window.location.pathname + window.location.search);
	} catch (err) {
		// ignore
	}

	e.preventDefault();
}

function collectHeadings(id) {
	const headings = [];

	(id ? document.getElementById(forRef) : document).querySelectorAll('h1, h2, h3, h4, h5, h6').forEach(heading => {
		headings.push({
			id: (n => {
				while (n) {
					if (n.id) {
						return n.id;
					}

					n = n.parentElement;
				}

				return undefined;
			})(heading),
			level: heading.tagName.toLowerCase(),
			text: heading.textContent.trim()
		});
	});

	return headings;
}


let forRef = null;

export { forRef };
export { buttonClass };

let headings = collectHeadings(forRef);

const levelClassMap = {
	'h1': 'font-semibold',
	'h2': 'pl-4',
	'h3': 'pl-8',
	'h4': 'pl-12',
	'h5': 'pl-16',
	'h6': 'pl-20',
};
</script>

<div class="relative z-0">
	<button
		class="relative z-10 group inline-block focus:outline-none cursor-pointer {buttonClass}"
		title="Table of Contents"
		on:click={handleButtonClick}
	>
		<div class="p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100 group-focus:pointer-events-none">
			{#if buttonActive}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
				</svg>
			{:else}
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5-3.9 19.5m-2.1-19.5-3.9 19.5" />
				</svg>
			{/if}
		</div>
	</button>

	{#if buttonActive}
		<!-- click bind for lazy propagate and close -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="z-0 fixed md:absolute left-1 top-10 md:top-0 w-screen md:max-w-sm"
			in:fly={{ x: -2, y: 0, duration: 100 }}
			out:fade={{ duration: 100 }}
			on:click={handleButtonClick}
		>
			<div class="bg-stone-800 md:ml-12 shadow-lg border-t md:border-t-0 border-stone-600 -translate-y-px md:translate-y-0">
				<div class="flex items-center justify-between mx-1.5 text-stone-400 border-b border-stone-600">
					<strong class="inline-block px-1.5 py-3 font-light font-lg">Table of Contents</strong>
					<a
						class="group h-full p-3 -mr-1.5 hover:text-white"
						href={window.location.toString().split('#')[0]}
						title="Scroll to top"
						aria-label="Scroll to top"
						on:click={scrollToTop}
					>
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pt-1 border-t border-stone-400 group-hover:border-white">
							<path fill-rule="evenodd" d="M10 17a.75.75 0 0 1-.75-.75V5.612L5.29 9.77a.75.75 0 0 1-1.08-1.04l5.25-5.5a.75.75 0 0 1 1.08 0l5.25 5.5a.75.75 0 1 1-1.08 1.04l-3.96-4.158V16.25A.75.75 0 0 1 10 17Z" clip-rule="evenodd" />
						</svg>
					</a>
				</div>
				<nav class="text-stone-200">
					<ol class="py-1">
						{#each headings as { id, level, text }}
							<li>
								<a class="block px-3 py-1 truncate hover:underline hover:text-white" href="#{id}">
									<span class="{levelClassMap[level]}">{text}</span>
								</a>
							</li>
						{/each}
					</ol>
				</nav>
			</div>
		</div>
		<button
			class="fixed inset-0 -z-10 bg-transparent"
			on:click={handleButtonClick}
			aria-label="Close Table of Contents"
		></button>
	{/if}
</div>
