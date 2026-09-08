<script>
  import { onDestroy } from "svelte";
  import ScriptLoader from "./ScriptLoader.svelte";

  const hlsAssetPath = __HLSJS_PATH__;
  const plyrAssetPath = __PLYR_PATH__;

  let playlistURL = null;
  let vttURL = null;
  let posterURL = null;

  export { playlistURL, vttURL, posterURL };

  let videoEl = null;
  let player = null;
  let hls = null;

  let hlsLoaded = false;
  let plyrLoaded = false;

  function initPlayer() {
    if (!hlsLoaded || !plyrLoaded || !videoEl || player) {
      return;
    }

    const Hls = window.Hls;
    const Plyr = window.Plyr;

    if (Hls && Hls.isSupported()) {
      // defer the actual manifest/segment fetching until playback is requested
      hls = new Hls({ autoStartLoad: false });
      hls.loadSource(playlistURL);
      hls.attachMedia(videoEl);
      videoEl.addEventListener('play', () => hls.startLoad(), { once: true });
    } else {
      // native HLS support (e.g. Safari); preload="none" keeps this lazy
      videoEl.src = playlistURL;
    }

    player = new Plyr(videoEl, {
      previewThumbnails: vttURL ? { enabled: true, src: vttURL } : undefined,
      controls: [
        'play-large',
        'play',
        'progress',
        'current-time',
        'mute',
        'volume',
        'captions',
        'settings',
        // 'pip', <- Remove or comment out this line
        'airplay',
        'fullscreen'
      ]
    });
  }

  function handleHlsLoad() {
    hlsLoaded = true;
    initPlayer();
  }

  function handlePlyrLoad() {
    plyrLoaded = true;
    initPlayer();
  }

  onDestroy(() => {
    if (player) player.destroy();
    if (hls) hls.destroy();
  });
</script>

<ScriptLoader url="{hlsAssetPath}/dist/hls.min.js" on:load={handleHlsLoad} />
<ScriptLoader url="{plyrAssetPath}/dist/plyr.min.js" cssUrl="{plyrAssetPath}/dist/plyr.css" on:load={handlePlyrLoad} />

<div class="h-full w-full">
  <video bind:this={videoEl} playsinline controls disablepictureinpicture crossorigin="anonymous" preload="none" poster={posterURL}></video>
</div>

<style>
  div :global(.plyr) {
    height: 100%;
    width: 100%;
  }
</style>
