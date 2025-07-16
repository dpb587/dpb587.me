<svelte:options customElement={{
	tag: 'dpb587-openseadragon',
	shadow: 'none',
	props: {
		infoURL: { type: 'String', attribute: 'info-url' },
		buttonClass: { type: 'String', attribute: 'button-class' }
	}
}} />

<script>
  import Interactive from "./Interactive.svelte";

  let buttonClass = "";
  let buttonActive = false;
  let infoURL = null;
  let infoDocument = null;
  let infoDocumentPromise = null;

  function handleButtonMouseover() {
    if (infoDocumentPromise) {
      return;
    }

    infoDocumentPromise = fetch(infoURL)
      .then(response => {
        if (!response.ok) {
          throw new Error(`Failed to fetch info document: ${response.statusText}`);
        }

        return response.json();
      })
      .then(data => {
        infoDocument = {
          ...data,
          // slight ux optimization to preload from mouseover; but also...
          // currently using relative urls in info.json which needs to be absolutized before openseadragon; should probably be absolute for max compatibility and external references
          'id': new URL(data['id'], window.location.toString()).toString(),
        };
      })
      .catch(error => {
        console.error("Error fetching info document:", error);
      })
  }

  function handleButtonClick(e) {
    buttonActive = !buttonActive;

    if (!buttonActive) {
      e.currentTarget.blur();
    }
  }

  export { buttonClass, infoURL };
</script>

<button
  class="relative z-10 group inline-block focus:outline-none {buttonClass}"
  title="Image Viewer"
  on:click={handleButtonClick}
  on:mouseover={handleButtonMouseover}
  on:focus={handleButtonMouseover}
>
  <div class="p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100 group-focus:pointer-events-none">
    {#if buttonActive}
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
      </svg>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
        <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607ZM10.5 7.5v6m3-3h-6" />
      </svg>
    {/if}
  </div>
</button>

{#if buttonActive}
  <div class="fixed inset-0 z-0 overflow-none">
    <button 
      class="absolute -z-10 -inset-8 backdrop-blur-lg grayscale"
      on:click={handleButtonClick}
      aria-label="Close Image Viewer"
    ></button>
    <div class="absolute top-10 md:top-0 left-0 md:left-12 bottom-0 right-0">
      <Interactive config={{
        tileSources: [infoDocument],
      }} />
    </div>
  </div>
{/if}
