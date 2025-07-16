<script>
  import { onMount, createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  export let url;
  let script;

  onMount(async () => {
    script.addEventListener('load', () => {
      dispatch('load');
    })

    script.addEventListener('error', (event) => {
      dispatch('error');
    });
  });
</script>

<svelte:head>
  <script bind:this={script} src={url}></script>
</svelte:head>
