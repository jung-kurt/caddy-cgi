# cgi [addon][addon]

[![MIT licensed][badge-mit]][license]
[![Build Status][badge-build]][travis]
[![Report][badge-report]][report]

> %addon-message%
> This directive is a Caddy extension. To get it, select this feature when you
> download Caddy. Questions should be directed to its maintainer.
> [github.com/jung-kurt/caddy-cgi][github]

Package cgi implements the common gateway interface ([CGI][cgi-wiki]) for
[Caddy][caddy], a modern, full-featured, easy-to-use web server.

This plugin lets you generate dynamic content on your website by means of
command line scripts. To collect information about the inbound HTTP request,
examine certain environment variables such as `PATH_INFO` and `QUERY_STRING`.
Then, to return a dynamically generated web page to the client, simply write
content to standard output. In the case of POST requests, you read additional
inbound content by means of the standard input.

The advantage of CGI is that you do not need to fuss with persistent server
startup, long term memory management, sockets, and crash recovery. Your script
is called when a request matches one the patterns that you specify in your
Caddyfile. As soon as your script completes its response it terminates. This
simplicity makes CGI a perfect complement to the straightforward operation and
configuration of Caddy. The benefits of Caddy, including HTTPS by default,
basic access authentication, and lots of middleware options extend easily to
your CGI scripts.

The disadvantage of CGI is that Caddy needs to start a new process for each
request. This could adversely impact your server's responsiveness in some
circumstances, such as when your web server is hit with very high demand, when
your script's dependencies require a long startup, or when concurrently running
scripts take a long time to respond. However, in many cases, such as using a
pre-compiled CGI application like fossil or a Lua script, the impact will
generally be insignificant.

> %block%
> **Important**: CGI scripts should be located outside of Caddy's document root.
> Otherwise, an inadvertent misconfiguration could result in Caddy delivering
> the script as an ordinary static resource. At best, this could merely confuse
> the site visitor. At worst, it could expose sensitive internal information
> that should not leave the server.

### Basic Syntax

The cgi directive lets you associate one or more patterns with a particular
script. The directive can be repeated any reasonable number of times. Here is
the basic syntax:

> %code%
> cgi [*match*][dir] [*exec*][arg] [[*args*][arg]...]

For example:

	cgi /report /usr/local/cgi-bin/report

When a request such as https://example.com/report or
https://example.com/report/weekly arrives, the cgi middleware will detect the
match and invoke the script named report that resides in the /usr/local/cgi-bin
directory. Here, it is assumed that the script is self-contained, for example a
pre-compiled CGI application or a shell script. An example of a standalone
script, similar to one used in the cgi plugin's test suite, follows:

	#!/bin/bash

	printf "Content-type: text/plain\n\n"
	printf "[%s %s %s %s %s]\n" $PATH_INFO $CGI_LOCAL $CGI_GLOBAL $1 $QUERY_STRING
	exit 0

The environment variables `PATH_INFO` and `QUERY_STRING` are populated and
passed to the script automatically. There are a number of other standard CGI
variables included that are described below. If you need to pass any special
environment variables or allow any environment variables that are part of
Caddy's process to pass to your script, you will need to use the advanced
directive syntax described below.

Fields that follow the exec directive are subject to placeholder replacement.
In addition to the standard Caddy placeholders such as `{method}` and `{host}`,
the following placeholders substitutions are made:

* **{.}** is replaced with Caddy's current working directory
* **{match}** is replaced with the portion of the request that satisfied the match
  directive
* **{root}** is replaced with Caddy's specified root directory

You can include glob wildcards in your matches. See the documentation for
[path/Match][match] in the Go standard library for more details about glob
matching. Here is an example directive:

	cgi /report/*.lua /usr/bin/lua /usr/local/cgi-bin/{match}

In this case, the cgi middleware will match requests such as
https://example.com/report/weekly.lua and
https://example.com/report/report.lua/weekly but not
https://example.com/report.lua. The use of the asterisk expands to any
character sequence within a directory. For example, if the request

	https://report/weekly.lua/summary

is made, the following command is executed:

	/usr/bin/lua /usr/local/cgi-bin/report/weeky.lua

Note that the portion of the request that follows the match is not included.
That information is conveyed to the script by means of environment variables.
In this example, the Lua interpreter is invoked directly from Caddy, so the Lua
script does not need the shebang that would be needed in a standalone script.
This method facilitates the use of CGI on the Windows platform.

### Advanced Syntax

In order to specify custom environment variables or pass along the environment
variables known to Caddy, you will need to use the advanced directive syntax.
That looks like this:

> %code%
> [cgi][dir] {
>   [app][subdir] [*match*][arg] [*script*][arg] [[*args*][arg]...]
>   [env][subdir] [*key1=val1*][arg] [[*key2=val2*][arg]...]
>   [pass_env][subdir] [*key1*][arg] [[*key2*][arg]...]
> }

Each of the keywords app, env, and pass_env may be repeated. The env and
pass_env lines are optional. If you wish to control environment variables at
the application level, the following syntax can be used:

> %code%
> [cgi][dir] {
>   [app][arg] {
>     [match][subdir] [*script*][arg] [[*args*][arg]...]
>     [env][subdir] [*key1=val1*][arg] [[*key2=val2*][arg]...]
>     [pass_env][subdir] [*key1*][arg] [[*key2*][arg]...]
>   }
>   [env][subdir] [*key1=val1*][arg] [[*key2=val2*][arg]...]
>   [pass_env][subdir] [*key1*][arg] [[*key2*][arg]...]
> }

### Environment Variable Example

In this example, the Caddyfile looks like this:

	192.168.1.2:8080
	root /usr/local/www
	cgi /show /usr/local/cgi-bin/report/gen

Note that a request for /show gets mapped to a script named
/usr/local/cgi-bin/report/gen. There is no need for any element of the script
name to match any element of the match pattern.

The contents of /usr/local/cgi-bin/report/gen are:

	#!/bin/bash

	printf "Content-type: text/plain\n\n"

	printf "example error message\n" > /dev/stderr

	if [ "POST" = "$REQUEST_METHOD" -a -n "$CONTENT_LENGTH" ]; then
	  read -n "$CONTENT_LENGTH" POST_DATA
	fi

	printf "AUTH_TYPE         [%s]\n" $AUTH_TYPE
	printf "CONTENT_LENGTH    [%s]\n" $CONTENT_LENGTH
	printf "CONTENT_TYPE      [%s]\n" $CONTENT_TYPE
	printf "GATEWAY_INTERFACE [%s]\n" $GATEWAY_INTERFACE
	printf "PATH_INFO         [%s]\n" $PATH_INFO
	printf "PATH_TRANSLATED   [%s]\n" $PATH_TRANSLATED
	printf "POST_DATA         [%s]\n" $POST_DATA
	printf "QUERY_STRING      [%s]\n" $QUERY_STRING
	printf "REMOTE_ADDR       [%s]\n" $REMOTE_ADDR
	printf "REMOTE_HOST       [%s]\n" $REMOTE_HOST
	printf "REMOTE_IDENT      [%s]\n" $REMOTE_IDENT
	printf "REMOTE_USER       [%s]\n" $REMOTE_USER
	printf "REQUEST_METHOD    [%s]\n" $REQUEST_METHOD
	printf "SCRIPT_EXEC       [%s]\n" $SCRIPT_EXEC
	printf "SCRIPT_NAME       [%s]\n" $SCRIPT_NAME
	printf "SERVER_NAME       [%s]\n" $SERVER_NAME
	printf "SERVER_PORT       [%s]\n" $SERVER_PORT
	printf "SERVER_PROTOCOL   [%s]\n" $SERVER_PROTOCOL
	printf "SERVER_SOFTWARE   [%s]\n" $SERVER_SOFTWARE

	exit 0

The purpose of this script is to show how request information gets communicated
to a CGI script. Note that POST data must be read from standard input. In this
particular case, posted data gets stored in the variable `POST_DATA`. Your
script may use a different method to read POST content. Secondly, the
`SCRIPT_EXEC` variable is not a CGI standard. It is provided by this middleware
and contains the entire command line, including all arguments, with which the
CGI script was executed.

When a browser requests

	http://192.168.1.2:8080/show/weekly?mode=summary

the response looks like

	AUTH_TYPE         []
	CONTENT_LENGTH    []
	CONTENT_TYPE      []
	GATEWAY_INTERFACE [CGI/1.1]
	PATH_INFO         [/weekly]
	PATH_TRANSLATED   []
	POST_DATA         []
	QUERY_STRING      [mode=summary]
	REMOTE_ADDR       [192.168.1.35]
	REMOTE_HOST       [192.168.1.35]
	REMOTE_IDENT      []
	REMOTE_USER       []
	REQUEST_METHOD    [GET]
	SCRIPT_EXEC       [/usr/local/cgi-bin/report/gen]
	SCRIPT_NAME       [/show]
	SERVER_NAME       [192.168.1.2:8080]
	SERVER_PORT       [8080]
	SERVER_PROTOCOL   [HTTP/1.1]
	SERVER_SOFTWARE   [go]

When a client makes a POST request, such as with the following command

	wget -O - -q --post-data="city=San%20Francisco" http://192.168.1.2:8080/show/weekly?mode=summary

the response looks the same except for the following lines:

	POST_DATA         [city=San%20Francisco]
	REQUEST_METHOD    [POST]

### Fossil Example

The fossil distributed software management tool is a native executable that
supports interaction as a CGI application. In this example, /usr/bin/fossil is
the executable and /home/quixote/projects.fossil is the fossil repository. To
configure Caddy to serve it, use a cgi directive something like this in your
Caddyfile:

	cgi /projects /usr/bin/fossil /usr/local/cgi-bin/projects

In your /usr/local/cgi-bin directory, make a file named projects with the
following single line:

	repository: /home/quixote/projects.fossil

The fossil documentation calls this a command file. When fossil is invoked
after a request to /projects, it examines the relevant environment variables
and responds as a CGI application.

[addon]: class:tag
[arg]: class:hl-arg
[badge-build]: https://travis-ci.org/jung-kurt/caddy-cgi.svg?branch=master
[badge-mit]: https://img.shields.io/badge/license-MIT-blue.svg
[badge-report]: https://goreportcard.com/badge/github.com/jung-kurt/caddy-cgi
[caddy]: https://caddyserver.com/
[cgi-wiki]: https://en.wikipedia.org/wiki/Common_Gateway_Interface
[dir]: class:hl-directive
[github]: https://github.com/jung-kurt/caddy-cgi
[license]: https://raw.githubusercontent.com/jung-kurt/caddy-cgi/master/LICENSE
[match]: https://golang.org/pkg/path/#Match
[report]: https://goreportcard.com/report/github.com/jung-kurt/caddy-cgi
[subdir]: class:hl-subdirective
[travis]: https://travis-ci.org/jung-kurt/caddy-cgi
