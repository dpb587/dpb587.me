<script>
  import { onDestroy, onMount } from "svelte";
    import ScriptLoader from "./ScriptLoader.svelte";

  let domContainer = null;
  let config = {};

  export { config };

  let viewer = null;

  function initViewer() {
    const OpenSeadragon = window.OpenSeadragon;

    viewer = OpenSeadragon({
      prefixUrl: '/assets/openseadragon/openseadragon-flat-toolbar-icons-current/images/',
      blendTime: 0.1,
      element: domContainer,
      immediateRender: true,
      minPixelRatio: 1,
      ...config,
    });
  }

  onMount(async() => {
    if (window.OpenSeadragon) {
      initViewer();
    }
  });

  onDestroy(async () => {
    if (viewer) {
      viewer.destroy();
    }
  });
</script>

<ScriptLoader url="/assets/openseadragon/openseadragon-current/openseadragon.min.js" on:load={initViewer} />
<div class="h-full w-full" bind:this={domContainer}></div>
