---
title: 3D Printing, Day 7
description: Hot Air Ballon; SF; deflectors; dry pods; OpenSCAD
publishDate: 2026-08-17
params:
  topics:
    3d-printing/journal: {}
---

Before the more interesting prints, some random progress from the weekend...

* I ran my first drying cycle of the 4 PLA Matte spools I had using the AMS 2 Pro.
* Moved around furniture and printer placement in preparation for a [workbench-style table](https://www.amazon.com/dp/B0FB8DM47R) that I ordered.
* Added some of the printed Inland filament tires to spools. One spool seemed slightly larger than standard and ended up splitting a printed tire trying to get it on.
* Ordered supplies for keeping filament dry. New Mexico is typically a dry climate, so it's less of a concern than many other areas of the country. I got a [pack of hygrometers](https://www.amazon.com/dp/B0DXCSXT9X) to see humidity and a bottle of [alumina dessicant beads](https://www.amazon.com/dp/B0B81VM24F) that I can keep alongside filament.

## AMS HT Installed

Finally hooked up the AMS HT to the printer. It needed a firmware upgrade, and then I ran a drying cycle for the PETG spool.

## Hot Air Balloon Print

A theme of Albuquerque is its morning hot air balloons (and [yearly festival](/topics/balloon-fiesta)), so I figured I should try one of the public hot air balloons designs. This was my first multi-plate project, each using its own color (all from the AMS 2 Pro).

* It seems like there was an odd slicing glitch which caused a short segment on the left half of the balloon to get mis-aligned.
* As it reached the top and was closing the top, the printing looked more fragile and surprisingly finished and looks reasonable.
* It seemed like the candle needed to be permanently snapped to the envelope, so I left those pieces separate for now.

{{< link-embed href="https://makerworld.com/en/models/1132818-hot-air-balloon-tea-light-candle#profileId-1133342" >}}

{{< video ref="./embed/media-p1g5nzpx7tn7.md" caption="Print Timelapse - Hot Air Balloon, Envelope" poster-frame="last" >}}

{{< video ref="./embed/media-n4xrnzvxj9f1.md" caption="Print Timelapse - Hot Air Balloon, Basket" poster-frame="last" >}}

{{< video ref="./embed/media-p5644t18j245.md" caption="Print Timelapse - Hot Air Balloon, Struts" poster-frame="last" >}}

## San Francisco City Print

Eventually, I want to learn how to print some of my own sections of cities, so San Francisco seemed like a good one to try out.

{{< link-embed href="https://makerworld.com/en/models/113500-san-francisco-california-3d-miniature#profileId-121972" >}}

{{< video ref="./embed/media-vd6q64b71c85.md" caption="Print Timelapse - San Francisco" poster-frame="last" >}}

## Fan Deflectors Print

I saw recommendations to print fan deflectors which help avoid strong, direct air current that can affect some printing. I printed them with the more heat-resistant PETG filament from the HT. This was my first print using the AMS HT, my first customization of a print in BambuStudio where I combined two profiles into a single plate, and my first print with support structures.

{{< link-embed href="https://makerworld.com/en/models/1968610-right-chamber-fan-deflector-x2d-p2s-no-warping#profileId-2116508" >}}
{{< link-embed href="https://makerworld.com/en/models/2675999-x2d-left-aux-fan-45deg-deflector-warping-denied#profileId-2962713" >}}

{{< video ref="./embed/media-gv1zrtk5vptk.md" caption="Print Timelapse - Fan Deflectors" poster-frame="last" >}}

## AMS Pro Dry Pods Print

The AMS Pro has a few spots for dessicant holders, so I printed some "dry pods" that can be kept in there. The fan deflectors were a super quick print with the HT/PETG, so I used this as a more thorough test of the HT and right nozzle. It definitely took a while though at 12 hours.

{{< link-embed href="https://makerworld.com/en/models/1534406-ams-2-pro-dry-pods-multiple-sizes-and-funnel#profileId-1609418" >}}

{{< video ref="./embed/media-py8ppbkp4xhn.md" caption="Print Timelapse - Dry Pods" poster-frame="last" >}}

## OpenSCAD

I did a little more research on [OpenSCAD](https://openscad.org/) and found that, although they haven't done a formal release in several years, it is still active enough and worth trying the nightly build. For some real-world, simpler objects, this code-based modeling seems like it may be a good match for some of the things I want to make.

As a quick experiment, I worked with code generation tools to see if it could create a hot air balloon model of my own with configurable gores, panels, size, and other parameters. Not quite a real-world design, but I was still impressed with how quickly it iterated something.

{{< image alt="OpenSCAD: Hot Air Balloon (Generated)" src="./media/openscad-hot-air-balloon.png" >}}

I exported/imported it into BambuStudio, but it lost color options and then wanted to add slice supports everywhere, so I have plenty more to experiment with.
