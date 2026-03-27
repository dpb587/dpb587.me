<script>
  export let sourceContentUri;
  export let sourceUrl;

  let content = null;
  let error = null;

  $: if (sourceContentUri) loadContent();

  function loadContent() {
    content = null;
    error = null;

    fetch(sourceContentUri)
      .then(r => {
        if (!r.ok) throw new Error(`Failed to fetch: ${r.statusText}`);
        return r.text();
      })
      .then(text => { content = text; })
      .catch(err => {
        error = err.message;
        console.error('Error fetching source:', err);
      });
  }

  function copyContent() {
    if (content !== null) {
      navigator.clipboard.writeText(content).catch(console.error);
    }
  }

  function downloadContent() {
    if (content === null) return;
    const blob = new Blob([content], { type: 'text/plain' });
    const objectUrl = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = objectUrl;
    a.download = 'source.md';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(objectUrl);
  }

  $: sourceHostname = (() => {
    try { return new URL(sourceUrl).pathname.split(/\/blob\/[^\/]+\//)[1]; } catch { return sourceUrl; }
  })();
</script>

<div class="md:px-12">
  <div class="flex items-center justify-between px-3 py-2 border-b border-stone-300 text-sm text-stone-800">
    <div class="flex items-center">
      <a
        href={sourceUrl}
        target="_blank"
        rel="noopener noreferrer"
        class="p-1 text-sm font-semibold text-stone-800 hover:text-black cursor-pointer underline decoration-stone-400"
      >{sourceHostname}</a>
    </div>
    <div class="flex items-center gap-1 -mr-2"></div>
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
