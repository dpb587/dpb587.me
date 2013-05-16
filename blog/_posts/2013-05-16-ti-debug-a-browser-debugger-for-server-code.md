---
title: "ti-debug: For Debugging Server Code in the Browser"
layout: post
tags: node xdebug webkit
description: Making it easier to debug languages like PHP and Python with only a browser.
---

I find that I am rarely using full IDEs to write code (e.g. [Eclipse][1], [Komodo][6], [NetBeans][3], [Zend Studio][2]).
They tend to be a bit sluggish when working with larger projects, so I favor simplistic editors like [Coda][4] or the
always-faithful [vim][5]. One thing I miss about using full-featured IDEs is their debugging capabilities. They usually
have convenient debugger interfaces that allow stepping through runtime code to investigate bugs.

About a year ago I started a project called [ti-debug][7] with the goal of being able to debug my server-side code (like
[PHP][8]) through [WebKit][13]'s developer tools interface. After getting a functional prototype of it working, I got
distracted with other projects and it dropped lower in the list of my repository activities. That is, until a few weeks
ago when [David][9] from [CityIndex][10] expressed interest in the project. I've been able to spend some sponsored time
in order to finish some of the features, update dependencies, and create a more stable project.


### Functionality

If you're familiar with the WebKit developer tools (also found in [Google Chrome][11]), the interface should look
extremely familiar. The core of `ti-debug` is written in [node.js][12] and when started up, it creates a simple web
server for you to open a browser tab and connect to. While you develop in other tabs, it will wait until there is an
incoming debug session at which point it loads up the debug environment and waits for you to step through code.

<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/waiting-to-debug.jpg"><img alt="Screenshot: waiting for connection" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/waiting-to-debug.jpg" width="308" /></a>
<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/initial-pause.jpg"><img alt="Screenshot: waiting for interaction" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/initial-pause.jpg" width="308" /></a>

The full stack trace is available along with all the local and global variables. In addition to the basic step
over/into/out, breakpoints can be set throughout the code. When paused, variables can be inspected and explored. In
addition to simple types like strings and booleans, complex objects and arrays can be expanded and further explored.

<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/breakpoints.jpg"><img alt="Screenshot: breakpoint exploration" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/breakpoints.jpg" width="628" /></a>

Not only can variables be read, they can also be updated inline by double clicking and entering new values. Or, for more
advanced commands, the console can be used to evaluate application code, possibly updating the runtime.

<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/propset-inline.jpg"><img alt="Screenshot: waiting for connection" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/propset-inline.jpg" width="308" /></a>
<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/propset-console.jpg"><img alt="Screenshot: waiting for interaction" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/propset-console.jpg" width="308" /></a>

Like most other IDE debuggers, the frontend supports jumping through the various levels in the stack to inspect the
runtime and run arbitrary commands. One other minor feature is watch expressions which are evaulated during every pause.

<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/stack-jumping.jpg"><img alt="Screenshot: waiting for connection" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/stack-jumping.jpg" width="308" /></a>
<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/watch-expressions.jpg"><img alt="Screenshot: waiting for interaction" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/watch-expressions.jpg" width="308" /></a>

Once a debug session has completed, the debug tab gets redirected back to the waiting page. Or, if the debug tab gets
closed in the middle of the debug session, the debugger will detach from the program and let it run to completion.

PHP isn't the only supported language. By using the debugging modules from [Komodo][14], other languages using the DBGp
communication can also use `ti-debug`. For example, Python scripts can currently be debugged, too...

<a href="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/python.jpg"><img alt="Screenshot: breakpoint exploration" src="/blog-data/2013-05-16-ti-debug-for-debugging-server-code-in-the-browser/python.jpg" width="628" /></a>


### Workflow

One of the ways that `ti-debug` can be run is locally for a single developer, but in the case of DBGp, `ti-debug` can
also act as a proxy to support multiple developers, or a combination of developers wanting to use both the browser-based
debugger along with their own local IDEs. This way, `ti-debug` could be running on a central development server to allow
all developers access.


 [1]: http://www.eclipse.org/
 [2]: http://www.zend.com/products/studio/
 [3]: https://netbeans.org/
 [4]: http://panic.com/coda/
 [5]: http://www.vim.org/
 [6]: http://www.activestate.com/komodo-ide
 [7]: https://github.com/dpb587/ti-debug
 [8]: http://php.net/
 [9]: https://github.com/mrdavidlaing
 [10]: https://github.com/cityindex
 [11]: https://www.google.com/intl/en/chrome/browser/
 [12]: http://nodejs.org/
 [13]: http://www.webkit.org/
 [14]: http://code.activestate.com/komodo/remotedebugging/
