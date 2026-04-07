<script>
  import { onMount } from 'svelte';

  export let structuredDataUri;

  const formats = [
    { value: 'jsonld', label: 'JSON-LD' },
    { value: 'ntriples', label: 'N-Triples' },
    { value: 'turtle', label: 'Turtle' },
  ];

  let format = 'jsonld';
  let content = null;
  let error = null;

  // Re-fetch whenever format or URI changes
  $: if (structuredDataUri) loadContent(format);

  function buildUrl(fmt) {
    const url = new URL(structuredDataUri, window.location.href);
    url.searchParams.set('format', fmt);
    return url.toString();
  }

  function loadContent(fmt) {
    content = null;
    error = null;

    fetch(buildUrl(fmt))
      .then(r => {
        if (!r.ok) throw new Error(`Failed to fetch: ${r.statusText}`);
        return r.text();
      })
      .then(text => {
        if (fmt === 'jsonld') {
          try { content = JSON.stringify(JSON.parse(text), null, 2); } catch { content = text; }
        } else {
          content = text;
        }
      })
      .catch(err => {
        error = err.message;
        console.error('Error fetching structured data:', err);
      });
  }

  function copyContent() {
    if (content !== null) {
      navigator.clipboard.writeText(content).catch(console.error);
    }
  }

  function downloadContent() {
    if (content === null) return;
    const extensions = { 'jsonld': 'jsonld', 'ntriples': 'nt', 'turtle': 'ttl' };
    const ext = extensions[format] ?? 'txt';
    const blob = new Blob([content], { type: 'text/plain' });
    const objectUrl = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = objectUrl;
    a.download = `structured-data.${ext}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(objectUrl);
  }
</script>

<div class="md:px-12">
  <div class="flex items-center justify-between px-3 py-2 border-b border-stone-300 text-sm text-stone-800">
    <div class="flex items-center">
      <span>View as</span>{' '}
      <select
        class="appearance-none bg-none bg-transparent border-0 p-1 text-sm font-semibold text-stone-800 hover:text-black cursor-pointer underline decoration-stone-400 focus:ring-0"
        bind:value={format}
      >
        {#each formats as f}
          <option value={f.value}>{f.label}</option>
        {/each}
      </select>
    </div>
    <div class="flex items-center gap-1 -mr-2">
      <button
        class="font-medium cursor-pointer px-2 py-1 hover:text-black hover:bg-neutral-200 disabled:cursor-default disabled:opacity-30 transition-colors"
        on:click={copyContent}
        disabled={content === null}
      >Copy</button>
      <button
        class="font-medium cursor-pointer px-2 py-1 hover:text-black hover:bg-neutral-200 disabled:cursor-default disabled:opacity-30 transition-colors"
        on:click={downloadContent}
        disabled={content === null}
      >Download</button>
      <a
        class="font-medium cursor-pointer px-2 py-1 hover:text-black hover:bg-neutral-200 transition-colors"
        href={buildUrl(format)}
        target="_blank"
        rel="noopener noreferrer"
      >Open</a>
    </div>
  </div>

  <div class="px-3 py-4">
    {#if error}
      <p class="text-red-600 text-sm">{error}</p>
    {:else if content === null}
      <p class="text-stone-400 text-sm">Loading…</p>
    {:else}
      <pre class="text-stone-800 text-sm whitespace-pre-wrap break-words font-mono leading-relaxed">{content}</pre>
    {/if}
  </div>
</div>
