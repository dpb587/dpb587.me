<script>
  import { onDestroy, onMount } from "svelte";
  import ScriptLoader from "./ScriptLoader.svelte";

  const assetPath = __PANNELLUM_PATH__;

  let domContainer = null;
  let basePath = null;
  let tileResolution = 512;
  let maxLevel = 4;
  let cubeResolution = 2048;

  export { basePath, tileResolution, maxLevel, cubeResolution };

  let viewer = null;
  let scriptLoaded = false;

  function initViewer() {
    if (!scriptLoaded || !domContainer) {
      return;
    }

    viewer = window.pannellum.viewer(domContainer, {
      type: "multires",
      autoLoad: true,
      autoRotate: -2,
      showControls: false,
      disableKeyboardCtrl: true,
      pitch: -16, // most are drone panoramas, so start with a downward angle
      multiRes: {
        basePath: basePath,
        path: "/%l/%s%y_%x",
        fallbackPath: "/fallback/%s",
        extension: "jpg",
        tileResolution: tileResolution,
        maxLevel: maxLevel,
        cubeResolution: cubeResolution,
      },
    });
  }

  onMount(() => {
    if (window.pannellum) {
      scriptLoaded = true;
      initViewer();
    }
  });

  onDestroy(() => {
    if (viewer) {
      viewer.destroy();
    }
  });

  function handleScriptLoad() {
    scriptLoaded = true;
    initViewer();
  }

  export function zoomIn() {
    if (viewer) viewer.setHfov(viewer.getHfov() - 10);
  }

  export function zoomOut() {
    if (viewer) viewer.setHfov(viewer.getHfov() + 10);
  }

  export function panLeft() {
    if (viewer) viewer.setYaw(viewer.getYaw() + 10);
  }

  export function panRight() {
    if (viewer) viewer.setYaw(viewer.getYaw() - 10);
  }

  export function panUp() {
    if (viewer) viewer.setPitch(viewer.getPitch() - 10);
  }

  export function panDown() {
    if (viewer) viewer.setPitch(viewer.getPitch() + 10);
  }
</script>

<ScriptLoader url="{assetPath}/pannellum.js" cssUrl="{assetPath}/pannellum.css" on:load={handleScriptLoad} />
<div class="h-full w-full" bind:this={domContainer}></div>
