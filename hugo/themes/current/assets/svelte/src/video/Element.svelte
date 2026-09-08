<svelte:options customElement={{
	tag: 'dpb587-video',
	shadow: 'none',
	props: {
		playlistURL: { type: 'String', attribute: 'playlist-url' },
		vttURL: { type: 'String', attribute: 'vtt-url' },
		posterURL: { type: 'String', attribute: 'poster-url' },
	}
}} />

<script>
  import Interactive from "./Interactive.svelte";

  let playlistURL = null;
  let vttURL = null;
  let posterURL = null;

  export { playlistURL, vttURL, posterURL };
</script>

<div class="h-full w-full">
  <Interactive playlistURL={playlistURL} vttURL={vttURL} posterURL={posterURL} />
</div>
