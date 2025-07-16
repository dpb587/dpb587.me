<svelte:options customElement={{
	tag: 'dpb587-elevation-chart',
	shadow: 'none',
	props: {
		src: { type: 'String', attribute: 'src' },
		containerClass: { type: 'String', attribute: 'container-class' }
	}
}} />

<script>
  import ElevationChartSVG from './internal/ElevationChartSVG.svelte';

  let containerClass;
  let src;

  export {
    containerClass,
    src,
  };

  let loader = (async () => {
    return {
      data: await fetch(src).then(async d => await d.json()),
    };
  })();
</script>

<div class={containerClass}>
  {#await loader}
    <div class="flex w-full h-full items-center justify-center text-neutral-400 uppercase tracking-tight font-mono text-xs ring-inset ring-1 ring-neutral-100">
      <span>Loading</span>
    </div>
  {:then svgProps}
    <ElevationChartSVG {...svgProps} />
  {:catch err}
    <div class="flex w-full h-full items-center justify-center text-neutral-600 uppercase tracking-tight font-mono font-bold text-xs shadow-inner">
      {'{'}<span class="text-red-800 cursor-help p-1 -mb-px" title={`Error: ${err.message}`}>Error</span>{'}'}
    </div>
  {/await}
</div>
