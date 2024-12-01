# go4ignition
**_go4ignition_ is NOT a framework.** This must be distinctly understood or nothing wonderful can come from using it. By design, there are no abstractions between its code and your code, unless you create them
yourself, and there is no update mechanism. If you wish to update, you must run `git pull` and work through
the resulting merge conflicts. It is expected that you fork it, change what you don't like, delete what you don't need,
and add what's missing. Hopefully, this will help you get off the ground faster, similar to a framework,
without incurring the ongoing maintenance costs of a framework. This approach isn't for everyone. Some people prefer
using a framework and that's great. However, if you would prefer to avoid a framework but struggle, as most of us do,
with time constraints then _go4ignition_ might be exactly what you're looking for.

# reload.sh
**This script has only been tested on Linux.** This is the core of the developer experience. Running this script will
compile your application, run it, and watch the project directory for changes. When a file is modified it will kill
your application, re-compile it, and restart it. A round trip typically takes less than 2 seconds from the moment you
save a file in your editor.