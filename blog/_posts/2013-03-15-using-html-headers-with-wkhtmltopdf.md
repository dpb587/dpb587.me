---
title: Using HTML Headers with wkhtmltopdf
layout: post
tags: [ 'headers', 'wkhtmltopdf' ]
description: Experimenting with dynamic HTML headers for PDFs.
---

Preparing for my job search, I really wanted to somehow reuse the content from my [about][2] page for my
r&#233;sum&#233; instead of trying to also maintain the information in a Word/Google Drive file. Mac OS X has the
convenient capability to convert any print to a PDF which is helpful in creating a general print-specific stylesheet for
browsers, but it still had a few drawbacks. One of those drawbacks is headers - I expect to see them on even the
simplest professional documents. Having used [`wkhtmltopdf`][1] before, I knew it could be a solution.

I started by creating a simple [header file][3] to include my name, my website, document name, and page information. I
also created a new CSS class which would take care of hiding headers and footers since they just take up extra space and
are being replaced. By using a few extra arguments, `wkhtmltopdf` does a brilliant job at creating a professional
document:

    wkhtmltopdf \
      --print-media-type \
      --run-script 'document.body.className+=" alt-printarticle";' \
      --margin-left 8mm --margin-right 8mm --margin-top 20mm \
      --header-spacing 3 \
      --header-html 'http://localhost:4000/include/content/header-simple.html?doctitle=r%26%23233%3Bsum%26%23233%3B' \
      --title 'resume' \
      'http://localhost:4000/about.html' \
      resume.pdf

Once that was working, I applied a few other tricks to make the printout a bit nicer:

 * [`page-break-inside`][5] &ndash; to prevent specific lines from breaking across pages (e.g. keeping the job title and
   company lines together)
 * `a` tag styling &ndash; suppressing underlines and visual differences since they make less sense when printed on
   paper
 * `.screen-only` and `.print-only` classes &ndash; to show slightly different content when printing (e.g. showing
   company website addresses instead of a generically linked "website" that looks simpler on browser screens)

Finally, after a bit of experimenting, learning, and styling, I can now present a consistent r&#233;sum&#233; (and cover
letters, references, &hellip;) whether it's through [PDF file][4] or [web page][2]. When viewing as a PDF, it has the
added benefits of remaining interactive with embedded links.


 [1]: https://code.google.com/p/wkhtmltopdf/
 [2]: /about.html
 [3]: /include/content/header-simple.html
 [4]: http://static.dpb587.me/about.pdf
 [5]: https://developer.mozilla.org/en-US/docs/CSS/page-break-inside
