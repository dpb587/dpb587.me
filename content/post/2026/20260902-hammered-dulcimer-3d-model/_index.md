---
description: Some non-trivial modeling.
params:
  featuredMedia:
    ref: embed/media-h56tp6nf3cb1
  topics:
    hammered-dulcimer/journal: {}
    openscad: {}
publishDate: 2026-09-02
title: 3D Hammered Dulcimer Model
---

I'm starting to learn 3D modeling and, currently, working with [OpenSCAD](https://openscad.org/) - a declarative, code-based method for building models. A hammered dulcimer seemed like a good, non-trivial example to learn more about the software (and is a model I want to use later). I started with lots of millimeter measurements of my own Master Works 15/14 instrument to create the foundational shapes. Then I got a lot more experience figuring out rotations, object difference'ing, and cylinders.

{{< model ref="./embed/media-h56tp6nf3cb1.md" caption="Hammered Dulcimer 3D Model" >}}

It looks more interesting with multiple colors, but that didn't come through in the 3D model web version here.

So far, this has all been manually-crafted code, but next I want to add strings and those are probably going to be complicated. I found a [BOSL2](https://github.com/BelfrySCAD/BOSL2/) - a popular library which supports a lot of the positional attachments and intersection helpers that I'll need. I think I will start asking [GitHub Copilot](https://github.com/features/copilot) for AI help when I get back to it, though.
