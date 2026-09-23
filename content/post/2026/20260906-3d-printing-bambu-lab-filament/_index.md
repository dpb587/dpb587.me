---
title: New Bambu Lab Filament
description: New inventory; first print error messages
publishDate: 2026-09-06
params:
  topics:
    3d-printing/journal: {}
---

## Bambu Lab Order

I wanted to test out the Bambu Lab filaments, so I ordered a few more interesting colors, some translucents that I want to experiment with, a few more-interesting styles, and a track switch to eventually install.

* PETG Translucent - [Light Blue (32600)](https://us.store.bambulab.com/products/petg-translucent?id=42720023445640), [Orange (32300)](https://us.store.bambulab.com/products/petg-translucent?id=42720026656904), [Clear (32101)](https://us.store.bambulab.com/products/petg-translucent?id=42479468281992)
* PETG Basic - [Black(30105)](https://us.store.bambulab.com/products/petg-basic?id=703099511296864269)
* PLA Matte - [Mandarin Orange (11300)](https://us.store.bambulab.com/products/pla-matte?id=40489681846408), [Marine Blue (11600)](https://us.store.bambulab.com/products/pla-matte?id=40489682141320), [Lilac purple (11700)](https://us.store.bambulab.com/products/pla-matte?id=40489682108552), [Grass Green (11500)](https://us.store.bambulab.com/products/pla-matte?id=40489681977480), [Lemon Yellow (11400)](https://us.store.bambulab.com/products/pla-matte?id=40489681911944), [Scarlet Red (11200)](https://us.store.bambulab.com/products/pla-matte?id=40489682043016)
* PLA Glow - [Glow Green (15500)](https://us.store.bambulab.com/products/pla-glow?id=41558487859336)
* PLA Metal - [Cobalt Blue Metallic (13600)](https://us.store.bambulab.com/products/pla-metal?id=41002951770248)
* PLA Silk Multi-Color - [Dawn Radiance (13912)](https://us.store.bambulab.com/products/pla-silk-multi-color?id=601310725769670660)
* [Reusable Spool](https://us.store.bambulab.com/products/bambu-reusable-spool?id=625124848578560003)
* [Filament Track Switch](https://us.store.bambulab.com/products/filament-track-switch)

## AMS Motor is Overloaded {#ams-motor-is-overloaded}

While printing more [spool desiccant holders](https://makerworld.com/en/models/1193993-high-performance-spool-desiccant-container-holder#profileId-1214551) for my new Bambu Lab filaments, I got my first print error after a couple hours of printing:

> The AMS assist motor is overloaded. This could be due to entangled filament or a stuck spool.

This ended up taking a few iterations to resolve...

1. First I looked up the troubleshooting guide. First impression was the rollers seemed like they may have been inexplicably locked. Manually tested them and restarted.
2. Same error after about an hour. Apparent-locked rollers seems like a red herring and just a result of the typical filament rewind process. I removed the filament roll, unwound a little bit, rewound, reinstalled, let it reload, and restarted.
3. Error. Again. Gave up for the night.
4. Restarted in the morning, but this time decided to camp out next to it and watch.
5. Errored again, but this time I noticed the filament was coming from the edge of the spool. I removed the spool, manually unwound it much more than before, and reinstalled it. It didn't seem particularly tight.

Finally, the print finished without errors. Next time I'll just skip to un-/re-winding the spool manually a bit more than I initially thought.

## Book Page Holder

Not much extra time for printing, but I learned that a "page spreader" is a thing, so randomly tried one of those prints using PLA True Red from Inland.

{{< link-embed href="https://makerworld.com/en/models/1687004-book-page-holder-v3?from=search#profileId-1787755" >}}

{{< video ref="./embed/media-xybt97722bp6.md" caption="Print Timelapse - Book Page Holder" poster-frame="last" >}}
