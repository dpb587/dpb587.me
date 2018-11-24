---
date: 2018-11-23
title: Switching from Jekyll to Hugo
description: Fewer dependencies, better information architecture, something new.
tags:
- blog
- docker
- hugo
- jekyll
- migration
---

It has been a while since I spent any time on my [personal website][1] here, but, recently, I have a few projects and ideas looking for a place to be recorded. As part of revisiting the site, I decided it might be a good opportunity to switch from using [Jekyll][2] as the static site generator to using [Hugo][3]. Here are some of the motivations and small learnings from that process.


## Jekyll

I first started with [Jekyll][2] back in 2013 when I was looking for a simple way to publish posts without the need for managing a complex environment, like [WordPress][5]. Eventually I decided to use Jekyll since it had a popular ecosystem, supported exactly what I was trying to do, and GitHub would take care of hosting the site on a custom domain for me. 

Over time, I started wanting to do more things which Jekyll or [GitHub Pages][4] didn't support. This included things like HTTPS for custom domains ([now supported][7]), allowing custom plugins, supporting photo galleries, storing assets outside of the repository. To support those, I ended up self-generating and self-hosting the site. It works, but, despite it being mostly automated, it has quite a bit of overhead if it's to be properly maintained.

Part of the automation included a [Docker][6] image and Gemfile for using the correct Jekyll and gem dependencies to build the site. The combination made it easier to run consistently anywhere, but it was still a fairly heavy dependency to have and keep updated (i.e. GitHub had been showing a [vulnerability alert][9] for one of the gems for a while). As part of a general move to lighter dependencies, I thought there should be room for improvement.

Another semi-related component was frontend asset management (like CSS and JavaScript). I never implemented a full asset-building pipeline, but did have some [simple scripts][8] to help statically build the assets with cacheable paths. I did not want to spend time with a complex asset build system like [gulp][10], but it would be nice to have some simple, effective asset management.


## Hugo

Similar to Jekyll, [Hugo][3] is static site generator. It encourages you to focus on writing content, organizing the content, and maintaining websites easily. While Jekyll is Ruby and gem-based, Hugo is executed with a single, static binary with no other runtime requirements. The minimal dependencies definitely made it simpler when getting started and automating processes.

One of its core concepts is the idea of different [content types][10]. Rather than there just being a "blog post" type of content, you can have one type for posts, another type for photos. Each content type automatically has its own theme files for different styling and presentation (which I find simpler than Jekyll's limited layout support).

Another difference is Hugo's native support of [taxonomies][11] and [content sections][12]. Taxonomies can be used to organize content across many content types, and content sections are more of a primary section with a content type. I think this will make it much easier as I experiment with using this site for a few more types of content.

A major difference from Jekyll is its templating system. Hugo uses the standard Go [template package][13] for rendering, so, by necessity the templates are very simple. While these templates are much more limited than Jekyll's [ERB][14] templates, it does help avoid theme files become too complex and unreadable.

One feature that Hugo has is basic [asset management][15] for CSS and JavaScript. Having worked with several other asset managers in the past I know there can be some complex setup and configuration required, but Hugo's approach is extremely simple with sane defaults and the declarative nature of the templates. While my site does not have complex assets, it's nice to support some best practices through the assets.


## Switching

The process of switching to Hugo was fairly straightforward, although it took several attempts over the course of a few months. There were a few memorable learnings from the process though...

**Structure** -- since Hugo handles content types a little differently than Jekyll, I moved around most of the files to match Hugo's strict, path-based expectations. Conceptually, archetypes, taxonomies, and sections were straightforward, but it took some experimenting before deciding how I wanted to represent my own content. For example, should a photo gallery be a taxonomy or a section; should those photos be their own type or serialized to data for a gallery; or how to organize private project planning drafts in sections?

Terminology-wise, I did get confused a few times on when the singular vs plural form of types should be used. I think taxonomies should always be plural (even in the `content` directory if adding `_index.md`), but that can be confusing if you also have a regular content type (which is singular). It took a couple reads of the documentation pages and some trial and error before I worked it out.

**Styling** -- as part of the switch and refreshing the site, I switched to [Bulma](https://bulma.io/) for CSS. My reasons for using Bulma were not very technical - there are plenty of good frameworks out there. I picked it primarily because: it was something I was not already familiar with, it used a different style of class naming, it was easy to imagine how a simple site like this one might look, and it was not a framework that I recognized as being overused by many other sites I encounter.

Overall, it worked out for my simple visual goals of the site. After working with it, I don't think I enjoy its class naming conventions (e.g. `has-text-centered` and `is-italic`) because the phrasing and conjugations seem unnatural to me. I did appreciate the range of simplistic building block components that it provides.

**Deployment** -- in the Jekyll version of the site, I had a dedicated Docker image which was intended to build the site. This was convenient, but it easily got out of date because it was not regularly rebuilt. Typically I would end up running whatever version of Jekyll I had on my workstation since I did not want to deal with gem versions or running it within Docker. With Hugo and its single binary, it turned out to be a much simpler dependency to manage and automatically upgrade.

**Deferred Work** -- there were a few things that did not seem worth switching on the initial pass:

 * I intentionally changed some of the URLs while switching. I used the `aliases` front matter option for the posts, but didn't go through the effort of recording aliases for existing photos and gallery pages since they're less of a priority. If Google Webmasters makes much of a noise, I figure I can backfill it with a script later.
 * In general, I did not spend much time on the photo or gallery pages. I'm still figuring out what the purpose and intent of those pages should be. I appreciate being able to host galleries outside of social media platforms, but I'm not sure if it is still worth the effort.

**Random Annoyances** -- there were a few other minor frustrations that I noticed while switching:

 * Auto-rebuild does not always auto-rebuild. Sometimes when updating a theme file, one page would reflect the change but another would not be updated. The `--disableFastRender` flag eventually helped avoid stale pages.
 * There were a few differences in how Markdown was rendered. I needed to review some of the pre-formatted code sections and list formatting of posts.


## Meta-Content

It is rare for me to revisit previously-written content for the sake of reviewing it. As part of the switch though, I at least skimmed through pages while I was finding and manually fixing conversion issues. It was interesting to revisit the different problems I have solved and consider how those solutions played out over time. In general, it was a good reminder for me to write posts which mark specific curiosities and discoveries I encounter (even if the writing is not the highest quality). When looking back, it makes it easier to consider how and why my experiences evolve over time.


[1]: https://dpb587.me/
[2]: https://jekyllrb.com/
[3]: https://gohugo.io/
[4]: https://pages.github.com/
[5]: https://wordpress.org/
[6]: https://www.docker.com/
[7]: https://blog.github.com/2018-05-01-github-pages-custom-domains-https/
[8]: https://github.com/dpb587/dpb587.me/blob/59829d7b0d7cfddd686b486da56562e0c8200f10/ci/tasks/build-docroot/run.sh#L26
[9]: https://blog.github.com/2017-11-16-introducing-security-alerts-on-github/
[10]: https://gohugo.io/content-management/types/
[11]: https://gohugo.io/content-management/taxonomies/
[12]: https://gohugo.io/content-management/sections/
[13]: https://golang.org/pkg/html/template/
[14]: https://ruby-doc.org/stdlib-2.5.3/libdoc/erb/rdoc/ERB.html
[15]: https://gohugo.io/categories/asset-management
