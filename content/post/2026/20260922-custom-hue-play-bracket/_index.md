---
title: Hue Play Bracket
description: Custom angled bracket for my shelf.
publishDate: 2026-09-22
params:
  featuredMedia:
    ref: embed/media-x119ygpc1z1n
  topics:
    3d-printing/journal: {}
    openscad: {}
---

I have a few [Hue Play light bars](https://www.philips-hue.com/en-us/p/hue-white-and-color-ambiance-play-light-bar-single-pack/7820130U7) next to the wall, but rarely end up using them. They either point directly at the wall and are more like a spotlight, or directly up at the ceiling (and then are in the line of sight and distractingly bright). I always wanted them mostly-hidden, but positioned with an upward angle to highlight travel photos I have printed. New 3D printer + learning 3D modeling + weekend...

{{< image ref="./embed/media-x119ygpc1z1n.md" caption="Three Test Prints" >}}

## Design

This was a fairly straightforward design, so I didn't even need AI to help me. I continue to rely on [OpenSCAD](https://openscad.org/) to start my real-world models. I took some millimeter-precision measurements of the bookcase and light bar, and then coded up some shapes. I went through three iterations and prints before I was happy with it. It also was my first time using the *Trim* feature in Bambu Studio which helped me avoid printing the whole piece when I only needed to verify particular aspects of the bracket.

1. First, the light bar was angled incorrectly and tight; and the lip for the shelf was too shallow.
2. Fixed and then added a hole for the screw to securely attach it to the light bar.
3. Finally printed the whole size. I switched to PETG Black in case the light bar gets warm. I pessimistically scaled it up an extra percent in size, but that turned out to be unnecessary.

{{< model ref="./embed/media-ppcp31p9qpvr.md" caption="Custom Hue Play Bracket" >}}

{{< details summary="Source Code" >}}

{{< markdown >}}The OpenSCAD model I used...{{< /markdown >}}

{{< snippet dir="appendix/2026-09-22-custom-hue-play-bracket" file="hue-play-bracket.scad" >}}

{{< /details >}}

Once I was happy with it, I kicked off a full set and got them installed - no more distracting light bars and a nice line of accent lights along the wall. Eventually I will be able to move them to the upper shelf once I move things back into position. And don't mind the now-highlighted dust...

{{< image ref="./embed/media-gktq2cj9byct.md" caption="Angled Hue Play light bars" >}}
