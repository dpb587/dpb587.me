<svelte:options customElement={{
	tag: 'dpb587-pannellum',
	shadow: 'none',
	props: {
		basePath: { type: 'String', attribute: 'base-path' },
		tileResolution: { type: 'Number', attribute: 'tile-resolution' },
		maxLevel: { type: 'Number', attribute: 'max-level' },
		cubeResolution: { type: 'Number', attribute: 'cube-resolution' },
		buttonClass: { type: 'String', attribute: 'button-class' }
	}
}} />

<script>
  import { fade } from 'svelte/transition';
  import Interactive from "./Interactive.svelte";

  let buttonClass = "";
  let buttonActive = false;
  let basePath = null;
  let tileResolution = 512;
  let maxLevel = 4;
  let cubeResolution = 2048;
  let interactiveRef = null;
  let overlayRef = null;
  let isFullscreen = false;

  function handleKeyUp(e) {
    switch (e.key) {
      case 'Escape': closeOverlay(); break;
      case '+': case '=': zoomIn(); break;
      case '-': zoomOut(); break;
      case 'f': case 'F': toggleFullscreen(); break;
      case 'ArrowLeft': if (interactiveRef) interactiveRef.panLeft(); break;
      case 'ArrowRight': if (interactiveRef) interactiveRef.panRight(); break;
      case 'ArrowUp': if (interactiveRef) interactiveRef.panUp(); break;
      case 'ArrowDown': if (interactiveRef) interactiveRef.panDown(); break;
    }
  }

  function handleFullscreenChange() {
    isFullscreen = !!document.fullscreenElement;
  }

  function closeOverlay() {
    buttonActive = false;
    document.body.classList.remove('overflow-hidden');
    document.removeEventListener('keyup', handleKeyUp);
    document.removeEventListener('fullscreenchange', handleFullscreenChange);
    if (document.fullscreenElement) {
      document.exitFullscreen();
    }
  }

  function handleButtonClick(e) {
    buttonActive = !buttonActive;
    if (buttonActive) {
      document.body.classList.add('overflow-hidden');
      document.addEventListener('keyup', handleKeyUp);
      document.addEventListener('fullscreenchange', handleFullscreenChange);
    } else {
      e.currentTarget.blur();
      closeOverlay();
    }
  }

  function zoomIn() {
    if (interactiveRef) interactiveRef.zoomIn();
  }

  function zoomOut() {
    if (interactiveRef) interactiveRef.zoomOut();
  }

  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      overlayRef?.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  }

  export { buttonClass, basePath, tileResolution, maxLevel, cubeResolution };
</script>

<button
  class="relative z-10 group block focus:outline-none {buttonClass}"
  title="Panorama Viewer"
  aria-label="Panorama Viewer"
  on:click={handleButtonClick}
>
  <div class="p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100 group-focus:pointer-events-none">
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5a17.92 17.92 0 0 1-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
    </svg>
  </div>
</button>

{#if buttonActive}
  <div
    class="fixed inset-0 z-60 md:z-40"
    bind:this={overlayRef}
    in:fade={{ duration: 100 }}
    out:fade={{ duration: 100 }}
  >
    <div class="fixed inset-0 -z-10 bg-stone-800"></div>
    <Interactive bind:this={interactiveRef} {basePath} {tileResolution} {maxLevel} {cubeResolution} />

    <div class="absolute top-0 left-0 flex text-stone-400">
      <button
        class="py-1 pl-1 pr-0.5 group inline-block focus:outline-none"
        on:click={closeOverlay}
        aria-label="Close"
        title="Close"
      >
        <div class="p-2 md:p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
          </svg>
        </div>
      </button>
      <button
        class="py-1 px-0.5 group inline-block focus:outline-none"
        on:click={zoomIn}
        aria-label="Zoom In"
        title="Zoom In (=, +)"
      >
        <div class="p-2 md:p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
        </div>
      </button>
      <button
        class="py-1 px-0.5 group inline-block focus:outline-none"
        on:click={zoomOut}
        aria-label="Zoom Out"
        title="Zoom Out (-)"
      >
        <div class="p-2 md:p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14" />
          </svg>
        </div>
      </button>
      <button
        class="py-1 pl-0.5 pr-1 group inline-block focus:outline-none"
        on:click={toggleFullscreen}
        aria-label={isFullscreen ? 'Exit Full Screen' : 'Full Screen'}
        title={isFullscreen ? 'Exit Full Screen (F)' : 'Full Screen (F)'}
      >
        <div class="p-2 md:p-2.5 bg-stone-800 group-hover:bg-black group-hover:text-white group-focus:ring-inset group-focus:ring-4 group-focus:ring-neutral-100">
          {#if isFullscreen}
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 9V4.5M9 9H4.5M9 9 3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5 5.25 5.25" />
            </svg>
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 md:w-5 h-4 md:h-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
            </svg>
          {/if}
        </div>
      </button>
    </div>
  </div>
{/if}
