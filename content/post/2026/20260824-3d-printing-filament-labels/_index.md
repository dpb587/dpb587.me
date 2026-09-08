---
title: 3D Printing, Filament Labels
description: Custom labels; Sandia topology
publishDate: 2026-08-24
params:
  topics:
    3d-printing/journal: {}
---

My new workbench arrived, so I got it assembled and moved the equipment on to it (and then ran a recalibration). I also added desiccant to the printed filament spool containers and installed them. Then I started learning more about OpenSCAD...

## Filament Labels

There are almost too many models on MakerWorld for labels and swatches, but none of them quite matched my preferences. I wanted something fairly simple (not articulated), medium size (not small/nano), label sticker with QR note links (not 3D-printed names; preferably for 30336 size I already use), and didn't require special attachments for keeping it with filament spools.

At some point, I found a [filament sample](https://makerworld.com/en/models/37649-filament-sample#profileId-36104) which included its OpenSCAD model file, so I was able to see how it worked and realize I could easily recreate my own layout. After some trial & error and a couple printed iterations, I ended up on a version that I think will work for me.

{{< image alt="Filament samples hanging on air-tight containers" src="./media/applied-labels.jpg" caption="Filament samples hanging on air-tight containers" >}}

{{< video ref="./embed/media-g446hp5y9fc7.md" caption="Print Timelapse - Filament Sample" poster-frame="last" >}}

{{< details summary="Source Code" >}}

{{< markdown >}}The OpenSCAD model I used...{{< /markdown >}}

{{< snippet dir="appendix/2026-08-24-3d-printing-filament-labels" file="FilamentSample.scad" >}}

{{< markdown >}}The basic HTML template I used...{{< /markdown >}}

{{< snippet dir="appendix/2026-08-24-3d-printing-filament-labels" file="Label-30336.html" lang="html" >}}

{{< /details >}}

## AMS 2 Pro Filament Tag Holder Print

I printed a tag holder for the AMS, reusing one of the public models which my swatches happen to fit. I used mostly the default settings based on the strong suggestion in the MakerWorld description, but it printed extremely slowly. I assume it was overly-configured for what seems like a simple print; next time I would look more closely.

{{< link-embed href="https://makerworld.com/en/models/3072960-clip-on-filament-tag-holder-for-ams-ams-2-pro#profileId-3514778" >}}

{{< video ref="./embed/media-tzc72chxkbbb.md" caption="Print Timelapse - AMS 2 Pro Filament Sample Holder" poster-frame="last" >}}

## Albuquerque Topology Print

I tried a topology print of my area - the Sandia Mountains and Rio Grande valley - that was listed publicly. It seemed like someone's experiment since it had some excess ground layers, so I tried some more of the BambuStudio tools to trim it. I also tried scaling the Z axis for slightly more exaggeration of the elevations.

{{< link-embed href="https://makerworld.com/en/models/1188696-albuquerque-and-sandia-mountain#profileId-1200103" >}}

{{< video ref="./embed/media-pv2v93h23g8q.md" caption="Print Timelapse - Albuquerque Topology" poster-frame="last" >}}

## ABS Galaxy Black

I ran a drying cycle for the ABS Galaxy Black (70º for 6hr). After it was complete, I was going to print a swatch sample, but got a warning that ABS apparently should not be used with the external extruder? Tried to move it inside the AMS 2 Pro, but the spool was slightly too large and couldn't rotate properly. Still need to figure out how it should normally be handled.
