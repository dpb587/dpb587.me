<svelte:options customElement={{
	tag: 'dpb587-code',
	shadow: 'none',
	props: {
		tilde: { type: 'String', attribute: 'tilde' },
		permalink: { type: 'String', attribute: 'permalink' },
		sourceUrl: { type: 'String', attribute: 'source-url' },
		buttonClass: { type: 'String', attribute: 'button-class' }
	}
}} />

<script>
  import { fade } from 'svelte/transition';
  import TextContent from './TextContent.svelte';
  import StructuredData from './StructuredData.svelte';
  import Source from './Source.svelte';

  let buttonClass = "";
  let buttonActive = false;
  let tilde = null;
  let permalink = null;
  let sourceUrl = null;
  let activeTab = 'textContent';
  let structuredDataMounted = false;
  let sourceMounted = false;

  $: base = tilde || '';
  $: textContentUri = permalink ? `${base}/~/export/text-content?uri=${permalink}` : null;
  $: structuredDataUri = permalink ? `${base}/~/export/structured-data?uri=${permalink}` : null;

  $: tabs = [
    { id: 'textContent', label: 'Text Content' },
    { id: 'structured-data', label: 'Structured Data' },
    ...(sourceUrl ? [{ id: 'source', label: 'Source' }] : []),
  ];

  $: sourceContentUri = sourceUrl
    ? `${base}/~/export/source?uri=${encodeURIComponent(sourceUrl)}`
    : null;

  // Lazily mount tabs on first visit
  $: if (activeTab === 'structured-data') structuredDataMounted = true;
  $: if (activeTab === 'source') sourceMounted = true;

  function handleEscKey(e) {
    if (e.key === 'Escape') closeOverlay();
  }

  function closeOverlay() {
    buttonActive = false;
    document.body.classList.remove('overflow-hidden');
    document.removeEventListener('keyup', handleEscKey);
    structuredDataMounted = false;
    sourceMounted = false;
    activeTab = 'textContent';
  }

  function handleButtonClick(e) {
    buttonActive = !buttonActive;
    if (buttonActive) {
      document.body.classList.add('overflow-hidden');
      document.addEventListener('keyup', handleEscKey);
    } else {
      e.currentTarget.blur();
      closeOverlay();
    }
  }

  function selectTab(id) {
    activeTab = id;
  }

  export { buttonClass, tilde, permalink, sourceUrl };
</script>

<button
  class="relative group inline-block focus:outline-none {buttonClass}"
  title="View Code"
  aria-label="View Code"
  on:click={handleButtonClick}
>
  <div class="p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100 group-focus:pointer-events-none">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
      <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" />
    </svg>
  </div>
</button>

{#if buttonActive}
  <div
    class="fixed inset-0 z-60 md:z-40 flex flex-col"
    in:fade={{ duration: 100 }}
    out:fade={{ duration: 100 }}
  >
    <!-- hide over-scroll content -->
    <div class="fixed inset-0 -z-10 bg-stone-50"></div>

    <div class="flex-none flex items-stretch bg-stone-800 text-stone-400">
      <button
        class="p-1 group inline-block focus:outline-none"
        on:click={handleButtonClick}
        aria-label="Close"
      >
        <div class="px-2.5 py-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
          </svg>
        </div>
      </button>
      <nav class="flex items-stretch">
        {#each tabs as tab}
          <button
            class="px-4 text-sm font-medium whitespace-nowrap focus:outline-none border-b-2 transition-colors
              {activeTab === tab.id
                ? 'text-white border-white'
                : 'text-stone-400 border-transparent hover:text-stone-200 hover:border-stone-500'}"
            on:click={() => selectTab(tab.id)}
          >
            {tab.label}
          </button>
        {/each}
      </nav>
    </div>

    <div class="flex-1 overflow-auto bg-stone-50">
      <div class:hidden={activeTab !== 'textContent'}>
        <TextContent textContentUri={textContentUri} />
      </div>
      {#if structuredDataMounted}
        <div class:hidden={activeTab !== 'structured-data'}>
          <StructuredData structuredDataUri={structuredDataUri} />
        </div>
      {/if}
      {#if sourceMounted}
        <div class:hidden={activeTab !== 'source'}>
          <Source sourceContentUri={sourceContentUri} sourceUrl={sourceUrl} />
        </div>
      {/if}
    </div>
  </div>
{/if}
