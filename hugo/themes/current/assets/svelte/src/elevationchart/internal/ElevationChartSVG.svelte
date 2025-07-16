<script>
  import * as d3 from 'd3';

  let domSVG = null;
  let data = null;

  export { data };

  const colorMap = {
    '#fafafa': '#0c4a6e', // text-sky-800
    '#d6d3d1': '#0f766e', // text-teal-700
    '#a8a29e': '#f59e0b', // text-amber-500
    '#57534e': '#ef4444', // text-red-500
  }

  function redraw() {
    const width = domSVG.clientWidth;
    const height = domSVG.clientHeight;
    const marginTop = 8;
    const marginRight = 8;
    const marginBottom = 25;
    const marginLeft = 44;
    
    // Create the scales.
    // const x = d3.scaleUtc()
    //     .domain(d3.extent(data, d => d.date))
    //     .rangeRound([marginLeft, width - marginRight]);
    const x = d3.scaleLinear()
    .domain(d3.extent(data, d => d.distance))
      .rangeRound([marginLeft, width - marginRight]);

    const y = d3.scaleLinear()
      .domain(d3.extent(data, d => d.elevation))
      .rangeRound([height - marginBottom, marginTop]);

    // const color = d3.scaleOrdinal(conditions.keys(), Array.from(conditions.values(), d => d.color))
    //   .unknown("white");

    // Create the path generator.
    const line = d3.line()
        .curve(d3.curveBasis)
        .x(d => x(d.distance))
        .y(d => y(d.elevation));

    // Create the SVG container.
    // const svg = d3.create("svg")
    const svg = d3.select(domSVG) //.append('svg')
        .attr("width", width)
        .attr("height", height)
        .attr("viewBox", [0, 0, width, height])
        .attr("style", "max-width: 100%; height: auto;")
        .on("pointerenter pointermove", pointermoved)
        .on("pointerleave", pointerleft)
        .on("touchstart", event => event.preventDefault());

      // svg.append("rect")
      // .attr("width", "100%")
      // .attr("height", "100%")
      // .attr("fill", "white");

    // Append the axes.
    svg.append("g")
        .attr("transform", `translate(0,${height - marginBottom})`)
        .call(d3.axisBottom(x).ticks(width / 80).tickSizeOuter(0))
        .call(g => g.select(".domain").remove());

    svg.append("g")
        .attr("transform", `translate(${marginLeft},0)`)
        .call(d3.axisLeft(y).ticks(Math.ceil(height / 40)))
        .call(g => g.select(".domain").remove())
        .call(g => g.select(".tick:last-of-type text").append("tspan").text(data.y));

    // Create the grid.
    // svg.append("g")
    //     .attr("stroke", "currentColor")
    //     .attr("stroke-opacity", 0.1)
    //     .call(g => g.append("g")
    //       .selectAll("line")
    //       .data(x.ticks())
    //       .join("line")
    //         .attr("x1", d => 0.5 + x(d))
    //         .attr("x2", d => 0.5 + x(d))
    //         .attr("y1", marginTop)
    //         .attr("y2", height - marginBottom))
    //     .call(g => g.append("g")
    //       .selectAll("line")
    //       .data(y.ticks())
    //       .join("line")
    //         .attr("y1", d => 0.5 + y(d))
    //         .attr("y2", d => 0.5 + y(d))
    //         .attr("x1", marginLeft)
    //         .attr("x2", width - marginRight));

    // Create the linear gradient.
    // const colorId = DOM.uid("color");
    // https://github.com/observablehq/stdlib/blob/main/src/dom/uid.js
    svg.append("linearGradient")
        .attr("id", 'TODO-unique-id')
        .attr("gradientUnits", "userSpaceOnUse")
        .attr("x1", 0)
        .attr("x2", width)
      .selectAll("stop")
      .data(data)
      .join("stop")
        .attr("offset", d => x(d.distance) / width)
        .attr("stop-color", d => colorMap[d.styleColor]); // color(d.condition));

    // Create the main path.
    svg.append("path")
        .datum(data)
        // .attr("fill", )
        .attr('fill', 'transparent')
        // .attr("stroke", '#292524') // text-stone-800
        .attr('stroke', "url(#TODO-unique-id)")
        .attr("stroke-width", 4)
        .attr("stroke-linejoin", "round")
        .attr("stroke-linecap", "round")
        .attr("d", line);

    const tooltip = svg.append("g");

    // Add the event listeners that show or hide the tooltip.
    const bisect = d3.bisector(d => d.distance).center;
    function pointermoved(event) {
      const i = bisect(data, x.invert(d3.pointer(event)[0]));
      tooltip.style("display", null);
      tooltip.attr("transform", `translate(${x(data[i].distance)},${y(data[i].elevation)})`);

      const path = tooltip.selectAll("path")
        .data([,])
        .join("path")
          .attr("fill", "#fafaf9") // text-stone-50
          .attr("stroke", "#292524"); // text-stone-800

      const text = tooltip.selectAll("text")
        .data([,])
        .join("text")
        .call(text => text
          .selectAll("tspan")
          .data([
            `${Math.round(data[i].distance*10)/10} km`,
            `Elevation, ${Math.round(data[i].elevation)} m`,
          ])
          .join("tspan")
            .attr("x", 0)
            .attr("y", (_, i) => `${i * 1.1}em`)
            .attr("font-weight", (_, i) => i ? null : "bold")
            .text(d => d));

      size(text, path);
    }

    function pointerleft() {
      tooltip.style("display", "none");
    }

    // Wraps the text with a callout path of the correct size, as measured in the page.
    function size(text, path) {
      const {x, y, width: w, height: h} = text.node().getBBox();
      text.attr("transform", `translate(${-w / 2},${15 - y})`);
      path.attr("d", `M${-w / 2 - 10},5H-5l5,-5l5,5H${w / 2 + 10}v${h + 20}h-${w + 20}z`);
    }
  }

  $: if (data && domSVG) {
    redraw();
  }
</script>

<svg class="h-full w-full" bind:this={domSVG} />
