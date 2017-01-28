/*
 * Copyright (c) 2017 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

/*
Package cgi implements the common gateway interface (CGI) for Caddy, a modern,
full-featured, easy-to-use web server.

CGI lets you generate dynamic content on your website by means of command line
scripts. To collect information about the inbound HTTP request, examine certain
environment variables such as PATH_INFO and QUERY_STRING. Then, to return a
dynamically generated web page to the client, simply write content to standard
output. In the case of POST requests, you read additional inbound content by
means of the standard input.

The advantage of CGI is that you do not need to fuss with persistent server
startup, long term memory management, sockets, and crash recovery. Your script
is called when a request matches one the patterns you specify in Caddyfile. As
soon as it completes its response it terminates. This simplicity makes CGI a
perfect complement to the straightforward operation and configuration of Caddy.
The benefits of Caddy, including HTTPS by default, basic access authentication,
and lots of middleware options extend easily to your CGI scripts.

The disadvantage of CGI is that Caddy needs to start a new process for each
request. This could adversely impact your server's responsiveness in some
circumstances, such as when your web server is hit with very high demand, when
your script's dependencies require a long startup, or when concurrently running
scripts take a long time to respond. However, in many cases, such as using a
pre-compiled CGI application like fossil or a Lua script, the impact will
generally be insignificant.

# Syntax

The cgi directive lets you associate one or more patterns with a particular
script. The directive can be repeated any reasonable number of times. Here is
the basic syntax:

```
cgi match exec [args...]
```

Here is an example.

```
cgi /report {root}/cgi-bin/report
```

When a request such as https://example.com/report or
https://example.com/report/weekly arrives, the cgi middleware will detect the
match and invoke the script named report that resides in the cgi-bin directory
which in turn belongs to the root directory configured in the Caddyfile. Here,
it is assumed that the script is self-contained, for example a pre-compiled CGI
application or a shell script. For example, the following script is similar to
one used in the cgi plugin's test suite:

```
#!/bin/bash

printf "Content-type: text/plain\n\n"
printf "[%s %s %s %s %s]\n" $PATH_INFO $CGI_LOCAL $CGI_GLOBAL $1 $QUERY_STRING
exit 0

```

The environment variables PATH_INFO and QUERY_STRING are populated and passed
to the script automatically. There are a number of other standard CGI variables
included that are described below. lIf you need to pass any special environment
variables or allow any environment variables that are part of Caddy's process
to pass to your script, you will need to use the advanced directive syntax
described below.

The exec field in the example includes the placeholder {root} that is
substituted with Caddy's specified root directory. You may also use the
placeholder {match} that will be substituted with the rooted script. In this
example, {match} would be /www/cgi-bin/report assuming that /www is Caddy's
configured root.

The optional arguments to the script can contain these placeholders as well as
the standard Caddy placeholders such as {method} and {host}.

You can include glob wildcards in your matches. Here is an example:

```
cgi /report/*.lua /usr/bin/lua {match}
```

In this case, the cgi middleware will match requests such as
https://example.com/report/weekly.lua and
https://example.com/report/report.lua/weekly but not
https://example.com/report.lua. The use of the asterisk expands to any
character sequence within a directory. The name of the matching script (it
could be something like /www/report/weekly.lua based on your Cadddyfile) will
be passed to the Lua interpreter. In this case, the Lua script does not need
the shebang that would be needed in a standalone script.

# Advanced syntax

In order to specify custom environment variables or pass along the environment
variables known to Caddy, you will need to use the advanced directive syntax.
That looks like this:

```
cgi {
  app match script [args...]
  env key1=val1 [keyn=valn...]
  pass_env key1 [keyn...]
}
```
Each of the keywords app, env, and pass_env may be repeated. The env and
pass_env lines are optional. If you wish to control environment variables at
the application level, the following syntax can be used:

cgi {
  app {
	  match script [args...]
	  env key1=val1 [keyn=valn...]
	  pass_env key1 [keyn...]
  }
  env key1=val1 [keyn=valn...]
  pass_env key1 [keyn...]
}

*/
package cgi
