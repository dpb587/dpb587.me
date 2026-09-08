<script>
  import { onMount, createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  export let url;
  export let cssUrl = null;
  let script;

  onMount(async () => {
    if (cssUrl) {
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = cssUrl;
      document.head.appendChild(link);
    }

    script.addEventListener('load', () => {
      dispatch('load');
    });

    script.addEventListener('error', (event) => {
      dispatch('error');
    });
  });
</script>

<svelte:head>
  <script bind:this={script} src={url}></script>
</svelte:head>
