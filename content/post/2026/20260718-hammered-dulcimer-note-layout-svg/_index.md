---
description: Generated diagrams of instrument notes
params:
  topics:
    hammered-dulcimer/journal: {}
publishDate: 2026-07-18
title: Note Layout SVG
---

The notes on a hammered dulcimer are often simplified as a boxes-and-lines diagram for reference. To help with some future learning aids, I made a React component to generate an SVG diagram. Since instruments have different sizes and occasionally different tunings, the layout is configurable. Eventually I'll use a React Context to help annotate and animate notes.

{{< image alt="15/14 Note Diagram" src="./media/screenshot.png" >}}

```jsx
<NoteDiagram instrument={{
  ...
  bridges: [
    {
      name: 'TREBLE',
      displayName: 'Treble',
      leftName: 'TL',
      rightName: 'TR',
      courses: [
        {
          leftNote: "D6",
          rightNote: "G5",
          stringCount: 2,
        },
        ...
      ],
    },
    {
      name: 'BASS',
      displayName: 'Bass',
      leftName: 'B',
      courses: [
        {
          leftNote: "C5",
          stringCount: 2,
        },
        ...
      ],
    }
  ]
}} />
```
