# CGI for Caddy

[![MIT licensed][badge-mit]][license]
[![Build Status][badge-build]][travis]
[![Report][badge-report]][report]

Package cgi implements the common gateway interface ([CGI][cgi-wiki]) for
[Caddy][caddy], a modern, full-featured, easy-to-use web server.

This plugin lets you generate dynamic content on your website by means of
command line scripts. To collect information about the inbound HTTP request,
your script examines certain environment variables such as
`PATH_INFO` and `QUERY_STRING`. Then, to return a dynamically
generated web page to the client, your script simply writes content to standard
output. In the case of POST requests, your script reads additional inbound
content from standard input.

The advantage of CGI is that you do not need to fuss with server startup and
persistence, long term memory management, sockets, and crash recovery. Your
script is called when a request matches one the patterns that you specify in
your Caddyfile. As soon as your script completes its response, it terminates.
This simplicity makes CGI a perfect complement to the straightforward operation
and configuration of Caddy. The benefits of Caddy, including HTTPS by default,
basic access authentication, and lots of middleware options extend easily to
your CGI scripts.

CGI has some disadvantages. For one, Caddy needs to start a new process for
each request. This can adversely impact performance and, if resources are
shared between CGI applications, may require the use of some interprocess
synchronization mechanism such as a file lock. Your server's responsiveness
could in some circumstances be affected, such as when your web server is hit
with very high demand, when your script's dependencies require a long startup,
or when concurrently running scripts take a long time to respond. However, in
many cases, such as using a pre-compiled CGI application like fossil or a Lua
script, the impact will generally be insignificant. Another restriction of CGI
is that scripts will be run with the same permissions as Caddy itself. This can
sometimes be less than ideal, for example when your script needs to read or
write files associated with a different owner.

### Security Considerations

Serving dynamic content exposes your server to more potential threats than
serving static pages. There are a number of considerations of which you should
be aware when using CGI applications.

**CGI scripts should be located outside of Caddy's document root.**
Otherwise, an inadvertent misconfiguration could result in Caddy delivering the
script as an ordinary static resource. At best, this could merely confuse the
site visitor. At worst, it could expose sensitive internal information that
should not leave the server.

**Mistrust the contents of `PATH_INFO`, `QUERY_STRING` and standard input.**
Most of the environment variables available to your CGI program are inherently
safe because they originate with Caddy and cannot be modified by external
users. This is not the case with `PATH_INFO`, `QUERY_STRING` and, in the case
of POST actions, the contents of standard input. Be sure to validate and
sanitize all inbound content. If you use a CGI library or framework to process
your scripts, make sure you understand its limitations.

### Application Modes

Your CGI application can be executed directly or indirectly. In the direct
case, the application can be a compiled native executable or it can be a shell
script that contains as its first line a shebang that identifies the
interpreter to which the file's name should be passed. Caddy must have
permission to execute the application. On Posix systems this will mean making
sure the application's ownership and permission bits are set appropriately; on
Windows, this may involve properly setting up the filename extension
association.

In the indirect case, the name of the CGI script is passed to an interpreter
such as lua, perl or python.

### Basic Syntax

The basic cgi directive lets you associate a single pattern with a particular
script. The directive can be repeated any reasonable number of times. Here is
the basic syntax:

	cgi match exec [args...]

For example:

	cgi /report /usr/local/cgi-bin/report

When a request such as https://example.com/report or
https://example.com/report/weekly arrives, the cgi middleware will detect the
match and invoke the script named /usr/local/cgi-bin/report. Here, it is
assumed that the script is self-contained, for example a pre-compiled CGI
application or a shell script. Here is an example of a standalone script,
similar to one used in the cgi plugin's test suite:

	#!/bin/bash

	printf "Content-type: text/plain\n\n"

	printf "PATH_INFO    [%s]\n" $PATH_INFO
	printf "QUERY_STRING [%s]\n" $QUERY_STRING

	exit 0

The environment variables `PATH_INFO` and `QUERY_STRING` are populated and
passed to the script automatically. There are a number of other standard CGI
variables included that are described below. If you need to pass any special
environment variables or allow any environment variables that are part of
Caddy's process to pass to your script, you will need to use the advanced
directive syntax described below.

The values used for the script name and its arguments are subject to
placeholder replacement. In addition to the standard Caddy placeholders such as
`{method}` and `{host}`, the following placeholders substitutions are made:

* **{.}** is replaced with Caddy's current working directory
* **{match}** is replaced with the portion of the request that satisfies the match directive
* **{root}** is replaced with Caddy's specified root directory

You can include glob wildcards in your matches. Basically, an asterisk
represents a sequence of zero or more non-slash characters and a question mark
represents a single non-slash character. These wildcards can be used multiple
times in a match expression. See the documentation for [path/Match][match] in
the Go standard library for more details about glob matching. Here is an
example directive:

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

In order to specify custom environment variables, pass along one or more
environment variables known to Caddy, or specify more than one match pattern
for a given rule, you will need to use the advanced directive syntax. That
looks like this:

	cgi {
		match match [match2...]
		exec script [args...]
		env key1=val1 [key2=val2...]
		pass_env key1 [key2...]
	}

For example,

	cgi {
		match /sample/*.php /sample/app/*.php
		exec /usr/local/cgi-bin/phpwrap /usr/local/cgi-bin{match}
		env DB=/usr/local/share/app/app.db SECRET=/usr/local/share/app/secret
		pass_env HOME UID
	}

With the advanced syntax, the `exec` subdirective must appear exactly once. The
`match` subdirective must appear at least once. The `env` and `pass_env`
subdirectives can appear any reasonable number of times.

The values associated with environment variable keys are all subject to
placeholder substitution, just as with the script name and arguments.

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

	CONTENT_LENGTH    [20]
	CONTENT_TYPE      [application/x-www-form-urlencoded]
	POST_DATA         [city=San%20Francisco]
	REQUEST_METHOD    [POST]

### Fossil Example

The [fossil][fossil] distributed software management tool is a native
executable that supports interaction as a CGI application. In this example,
/usr/bin/fossil is the executable and /home/quixote/projects.fossil is the
fossil repository. To configure Caddy to serve it, use a cgi directive
something like this in your Caddyfile:

	cgi /projects /usr/bin/fossil /usr/local/cgi-bin/projects

In your /usr/local/cgi-bin directory, make a file named projects with the
following single line:

	repository: /home/quixote/projects.fossil

The fossil documentation calls this a command file. When fossil is invoked
after a request to /projects, it examines the relevant environment variables
and responds as a CGI application. If you protect /projects with
[basic HTTP authentication][auth], you may wish to enable the
**Allow REMOTE_USER authentication** option when setting up fossil. This lets
fossil dispense with its own authentication, assuming it has an account for
the user.

### Agedu Example

The [agedu][agedu] utility can be used to identify unused files that are taking
up space on your storage media. Like fossil, it can be used in different modes
including CGI. First, use it from the command line to generate an index of a
directory, for example

	agedu --file /home/quixote/agedu.dat --scan /home/quixote

In your Caddyfile, include a directive that references the generated index:

	cgi /agedu /usr/local/bin/agedu --cgi --file /home/quixote/agedu.dat

You will want to protect the /agedu resource with some sort of access control,
for example [HTTP Basic Authentication][auth].

### Go Source Example

This small example demonstrates how to write a CGI program in Go. The use of a
bytes.Buffer makes it easy to report the content length in the CGI header.

	package main

	import (
		"bytes"
		"fmt"
		"os"
		"time"
	)

	func main() {
		var buf bytes.Buffer

		fmt.Fprintf(&buf, "Server time at %s is %s\n",
			os.Getenv("SERVER_NAME"), time.Now().Format(time.RFC1123))
		fmt.Println("Content-type: text/plain")
		fmt.Printf("Content-Length: %d\n\n", buf.Len())
		buf.WriteTo(os.Stdout)
	}

When this program is compiled and installed as /usr/local/bin/servertime, the
following directive in your Caddy file will make it available:

	cgi /servertime /usr/local/bin/servertime

### Cgit Example

The [cgit][cgit] application provides an attractive and useful web interface to
git repositories. Here is how to run it with Caddy. After compiling cgit, you
can place the executable somewhere out of Caddy's document root. In this
example, it is located in /usr/local/cgi-bin.

A sample configuration file is included in the project's cgitrc.5.txt file. You
can use it as a starting point for your configuration. The default location for
this file is /etc/cgitrc but in this example the location
/home/quixote/caddy/cgitrc. Note that changing the location of this file from
its default will necessitate the inclusion of the environment variable
CGIT_CONFIG in the Caddyfile cgi directive.

When you edit the repository stanzas in this file, be sure each repo.path item
refers to the .git directory within a working checkout. Here is an example
stanza:

	repo.url=caddy-cgi
	repo.path=/home/quixote/go/src/github.com/jung-kurt/caddy-cgi/.git
	repo.desc=CGI for Caddy
	repo.owner=jung-kurt
	repo.readme=/home/quixote/go/src/github.com/jung-kurt/caddy-cgi/README.md

Also, you will likely want to change cgit's cache directory from its default
in /var/cache (generally accessible only to root) to a location writeable by
Caddy. In this example, cgitrc contains the line

	cache-root=/home/quixote/.cache/cgit

You may need to create the cgit subdirectory.

There are some static cgit resources (namely, cgit.css, favicon.ico, and
cgit.png) that will be accessed from Caddy's document tree. For this example,
these files are placed in a directory named cgit-resource. The following lines
are part of the cgitrc file:

	css=/cgit-resource/cgit.css
	favicon=/cgit-resource/favicon.ico
	logo=/cgit-resource/cgit.png

Additionally, you will likely need to tweak the various file viewer filters
such source-filter and about-filter based on your system.

The following Caddyfile directive will allow you to access the cgit application
at /cgit:

	cgi {
		match /cgit
		exec /usr/local/cgi-bin/cgit
		env CGIT_CONFIG=/home/quixote/caddy/cgitrc
	}

### PHP Example

Feeling reckless? You can run [PHP][php] in CGI mode. In general,
[FastCGI][fastcgi] is the preferred method to run PHP if your application has
many pages or a fair amount of database activity. But for small PHP programs
that are seldom used, CGI can work fine. You'll need the php-cgi interpreter
for your platform. This may involve downloading the executable or downloading
and then compiling the source code. For this example, assume the interpreter is
installed as /usr/local/bin/php-cgi. Additionally, because of the way PHP
operates in CGI mode, you will need an intermediate script. This one works in
Posix environments:

	#!/bin/bash

	REDIRECT_STATUS=1 SCRIPT_FILENAME="${1}" /usr/local/bin/php-cgi -c /home/quixote/.config/php/php-cgi.ini

This script can be reused for multiple cgi directives. In this example, it is
installed as /usr/local/cgi-bin/phpwrap. The argument following -c is your
initialization file for PHP. In this example, it is named
/home/quixote/.config/php/php-cgi.ini.

Two PHP files will be used for this example. The first,
/usr/local/cgi-bin/sample/min.php, looks like this:

	<!DOCTYPE html>
	<html>
	  <head>
	    <title>PHP Sample</title>
	    <style>
	      form span {
	        font: 15px sans-serif;
	        display: inline-block;
	        width: 8em;
	        text-align: right;
	      }
	    </style>
	  </head>
	  <body>
	    <form action="action.php" method="post">
	      <p><span>Name</span> <input type="text" name="name" /></p>
	      <p><span>Number</span> <input type="text" name="number" /></p>
	      <p><span>Day</span> <input type="text" name="day"
	        value="<?php echo(date("l", time())); ?>" /></p>
	      <p><span>&nbsp;</span> <input type="submit" /></p>
	    </form>
	  </body>
	</html>

The second, /usr/local/cgi-bin/sample/action.php, follows:

	<!DOCTYPE html>
	<html>
	  <head>
	    <title>PHP Sample</title>
	  </head>
	  <body>
	    <p>Name is <strong><?php echo htmlspecialchars($_POST['name']); ?></strong>.</p>
	    <p>Number is <strong><?php echo (int)$_POST['number']; ?></strong>.</p>
	    <p>Day is <strong><?php echo htmlspecialchars($_POST['day']); ?></strong>.</p>
	  </body>
	</html>

The following directive in your Caddyfile will make the application available
at sample/min.php:

	cgi /sample/*.php /usr/local/cgi-bin/phpwrap /usr/local/cgi-bin{match}


[agedu]: http://www.chiark.greenend.org.uk/~sgtatham/agedu/
[auth]: https://caddyserver.com/docs/basicauth
[badge-build]: https://travis-ci.org/jung-kurt/caddy-cgi.svg?branch=master
[badge-mit]: https://img.shields.io/badge/license-MIT-blue.svg
[badge-report]: https://goreportcard.com/badge/github.com/jung-kurt/caddy-cgi
[caddy]: https://caddyserver.com/
[cgi-wiki]: https://en.wikipedia.org/wiki/Common_Gateway_Interface
[fossil]: https://www.fossil-scm.org/
[github]: https://github.com/jung-kurt/caddy-cgi
[license]: https://raw.githubusercontent.com/jung-kurt/caddy-cgi/master/LICENSE
[match]: https://golang.org/pkg/path/#Match
[report]: https://goreportcard.com/report/github.com/jung-kurt/caddy-cgi
[travis]: https://travis-ci.org/jung-kurt/caddy-cgi
[php]: http://php.net/
[fastcgi]: https://caddyserver.com/docs/fastcgi
[cgit]: https://git.zx2c4.com/cgit/about/