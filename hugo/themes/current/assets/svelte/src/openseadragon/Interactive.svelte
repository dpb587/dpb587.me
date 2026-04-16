<script>
  import { onDestroy, onMount } from "svelte";
    import ScriptLoader from "./ScriptLoader.svelte";

  const assetPath = __OPENSEADRAGON_PATH__;

  let domContainer = null;
  let config = {};

  export { config };

  let viewer = null;

  function initViewer() {
    const OpenSeadragon = window.OpenSeadragon;

    viewer = OpenSeadragon({
      blendTime: 0.1,
      element: domContainer,
      immediateRender: true,
      minPixelRatio: 1,
      defaultZoomLevel: 1.1,
      showNavigationControl: false,
      keyboardNavEnabled: false,
      gestureSettingsMouse: {
        clickToZoom: false,
        dblClickToZoom: true,
      },
      ...config,
    });
  }

  export function zoomIn() {
    if (viewer) viewer.viewport.zoomBy(1.5);
  }

  export function zoomOut() {
    if (viewer) viewer.viewport.zoomBy(1 / 1.5);
  }

  export function panLeft() {
    if (viewer) viewer.viewport.panBy(new window.OpenSeadragon.Point(0.1, 0));
  }

  export function panRight() {
    if (viewer) viewer.viewport.panBy(new window.OpenSeadragon.Point(-0.1, 0));
  }

  export function panUp() {
    if (viewer) viewer.viewport.panBy(new window.OpenSeadragon.Point(0, 0.1));
  }

  export function panDown() {
    if (viewer) viewer.viewport.panBy(new window.OpenSeadragon.Point(0, -0.1));
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

<ScriptLoader url="{assetPath}/openseadragon.min.js" on:load={initViewer} />
<div class="h-full w-full" bind:this={domContainer}></div>
