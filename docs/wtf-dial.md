Introducing WTF Dial (again)
============================

A blog series where we build and deploy a real-world Go application.

[

Ben Johnson

04 Jan 2021

](https://web.archive.org/web/20251012024039/https://www.gobeyond.dev/author/benbjohnson/)

*   [](https://web.archive.org/web/20251012024039/https://twitter.com/share?text=Introducing%20WTF%20Dial%20(again)&url=https://www.gobeyond.dev/wtf-dial/ "Share on Twitter")
*   [](https://web.archive.org/web/20251012024039/https://www.facebook.com/sharer/sharer.php?u=https://www.gobeyond.dev/wtf-dial/ "Share on Facebook")
*   [](https://web.archive.org/web/20251012024039/https://www.linkedin.com/shareArticle?mini=true&url=https://www.gobeyond.dev/wtf-dial/&title=Introducing%20WTF%20Dial%20(again) "Share on LinkedIn")
*   [](https://web.archive.org/web/20251012024039/mailto:/?subject=Introducing%20WTF%20Dial%20(again)&body=https://www.gobeyond.dev/wtf-dial/ "Share via Email")

![](https://web.archive.org/web/20251012024039im_/https://www.gobeyond.dev/content/images/2020/12/2020-12-28-12.50.50.gif)

Updating your WTF level in real-time at [wtfdial.com](https://web.archive.org/web/20251012024039/https://wtfdial.com/?ref=gobeyond.dev)

A little history
----------------

Several years ago I started a collaborative, open-source project called WTF Dial to provide an example of how to build a real project in Go. The idea was to build a slice of the application every week and write a post about it.

Turns out that's a bad idea.

If there's one thing that I learned from that experience it's that I don't build applications in a linear fashion. I'll start on the database layer, jump to writing an HTML frontend, then circle back to implement some HTTP handling, and maybe throw in a CLI somewhere in there. Releasing the application in incremental, clean slices was never going to happen.

Since abandoning the project, I've had many requests to finish it. So I started over and built the entire application first. Now I'll write about it. You can find the application at [https://wtfdial.com/](https://web.archive.org/web/20251012024039/https://wtfdial.com/?ref=gobeyond.dev) and you can find the [source code on GitHub](https://web.archive.org/web/20251012024039/https://github.com/benbjohnson/wtf?ref=gobeyond.dev).

Huge thanks to [Cory LaNou](https://web.archive.org/web/20251012024039/https://twitter.com/corylanou?ref=gobeyond.dev) from [Gopher Guides](https://web.archive.org/web/20251012024039/https://www.gopherguides.com/?ref=gobeyond.dev) for helping to review the project!

Ok, so WTF is WTF Dial?
-----------------------

It all started back in 2016 with a [series of tweets from Peter Bourgon](https://web.archive.org/web/20251012024039/https://twitter.com/peterbourgon/status/765935213507649537?ref=gobeyond.dev):

> "I have an idea for a thing which helps with remote work, it is called The WTF Dial, let me explain. It is some analog knob or dial which you can dial to zero (No WTFs, all clear) or 100% (WTF?!)  
>   
> The current state of the dial goes to some dashboard, along with all of your teammates. So at any moment you can see how fucked everyone is. Just as a passive indicator. If someone is WTF-ing, you can ask them if they want to chat, and help them get un-WTF'd."  
>   
> —[Peter Bourgon](https://web.archive.org/web/20251012024039/https://twitter.com/peterbourgon?ref=gobeyond.dev)

Essentially, it's a way to quantify and aggregate the feelings of your entire team into a single number. Sounds callous, sure, but it's just for fun.

Topics that we'll cover in this series
--------------------------------------

In this series, we'll start with architecture, move our way from the backend to the frontend, and also describe deployment strategies. We'll cover topics such as:

*   Application design & code structure
*   Working with SQL databases
*   HTTP & WebSockets
*   Using `context.Context` appropriately
*   Embedded file systems (introduced in Go 1.16)
*   Unit testing & end-to-end testing
*   Command line interfaces
*   CI/CD
*   Performance analysis & load testing

Have suggestions for topics you want covered or questions you want answered? Please [c](https://web.archive.org/web/20251012024039/mailto:ben@gobeyond.dev)heck out the [WTF Dial GitHub Discussions](https://web.archive.org/web/20251012024039/https://github.com/benbjohnson/wtf/discussions?ref=gobeyond.dev) page to chat.

Why Go?
-------

The Go programming language is known as a low-level language used for cloud infrastructure but I find it to be a fantastic language for general application development. It provides fast compile times for quick iteration that you'd expect from a language like Ruby but you also get basic type checking and performance like you'd find in a language like Java or C#.

The development experience with Go is great but it gets really good once you deploy. Applications are statically compiled so there is no runtime dependency or shared libraries to worry about. Just upload your binary and run it.

Go applications can have a small footprint as well so we can support hundreds of users with our application with even a small $5/month VPS. It's great for small applications like ours.

Follow along
------------

If you're interested in this series, please sign up for the newsletter so you can get each new post in your inbox. [Click here to subscribe for free.](https://web.archive.org/web/20251012024039/https://www.gobeyond.dev/signup/)

[WTF Dial](/web/20251012024039/https://www.gobeyond.dev/tag/wtf-dial/)

[](/web/20251012024039/https://www.gobeyond.dev/author/benbjohnson/)

### [Ben Johnson](/web/20251012024039/https://www.gobeyond.dev/author/benbjohnson/)[

](https://web.archive.org/web/20251012024039/https://x.com/benbjohnson)

Freelance Go developer, author of BoltDB

* * *
