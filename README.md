# cgi

[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/jung-kurt/caddy-cgi/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jung-kurt/caddy-cgi?status.svg)](https://godoc.org/github.com/jung-kurt/caddy-cgi)
[![Build Status](https://travis-ci.org/jung-kurt/caddy-cgi.svg?branch=master)](https://travis-ci.org/jung-kurt/caddy-cgi)
[![Report](https://goreportcard.com/badge/github.com/jung-kurt/caddy-cgi)](https://goreportcard.com/report/github.com/jung-kurt/caddy-cgi)

Package cgi implements the common gateway interface (CGI) for Caddy, a modern,
full-featured, easy-to-use web server.

This plugin lets you generate dynamic content on your website by means of
command line scripts. To collect information about the inbound HTTP request,
examine certain environment variables such as PATH_INFO and QUERY_STRING. Then,
to return a dynamically generated web page to the client, simply write content
to standard output. In the case of POST requests, you read additional inbound
content by means of the standard input.

The advantage of CGI is that you do not need to fuss with persistent server
startup, long term memory management, sockets, and crash recovery. Your script
is called when a request matches one the patterns you specify in your
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

##Basic Syntax
The cgi directive lets you associate one or more patterns with a particular
script. The directive can be repeated any reasonable number of times. Here is
the basic syntax:

```
cgi match exec [args...]
```

Here is an example:

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
included that are described below. If you need to pass any special environment
variables or allow any environment variables that are part of Caddy's process
to pass to your script, you will need to use the advanced directive syntax
described below.

The exec field in the example includes the placeholder {root} that is
substituted with Caddy's specified root directory. You may also use the
placeholder {match} that will be substituted with the rooted script name. In
this example, {match} would be /www/cgi-bin/report assuming that /www is
Caddy's configured root.

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
the shebang that would be needed in a standalone script. See the documentation
for path/Match in the Go standard library for more details about glob matching.

##Advanced Syntax
In order to specify custom environment variables or pass along the environment
variables known to Caddy, you will need to use the advanced directive syntax.
That looks like this:

```
cgi {
  app match script [args...]
  env key1=val1 [key2=val2...]
  pass_env key1 [key2...]
}
```

Each of the keywords app, env, and pass_env may be repeated. The env and
pass_env lines are optional. If you wish to control environment variables at
the application level, the following syntax can be used:

```
cgi {
  app {
    match script [args...]
    env key1=val1 [key2=val2...]
    pass_env key1 [key2...]
  }
  env key1=val1 [key2=val2...]
  pass_env key1 [key2...]
}
```

##Environment Variable Example
In this example, the Caddyfile looks like this:

```
192.168.1.2:8080
root /var/www
cgi /show {root}/report/gen
```

Note that a request for /show gets mapped to a script named
/var/www/report/gen. There is no need for any element of the script name to be
the same as the match pattern.

The script stored with the name /var/www/report/gen looks like this:

```
#!/bin/bash

printf "Content-type: text/plain\n\n"

printf "example error message\n" > /dev/stderr

if [ "POST" = $REQUEST_METHOD -a -n $CONTENT_LENGTH ]; then
  read -n $CONTENT_LENGTH POST_DATA
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
```

The purpose of this script is to show how request information gets communicated
to a CGI script. Note that POST data must be read from standard input. In this
particular case, posted data gets stored in the variable POST_DATA. Your script
may use a different method to read POST content. Secondly, the SCRIPT_EXEC
variable is not a CGI standard. It is provided by this middleware and contains
the entire command line, including all arguments, with which the CGI script was
executed.

When a browser requests

```
http://192.168.1.2:8080/show/weekly?mode=summary
```

the response looks like

```
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
SCRIPT_EXEC       [/var/www/report/gen]
SCRIPT_NAME       [/show]
SERVER_NAME       [192.168.1.2:8080]
SERVER_PORT       [8080]
SERVER_PROTOCOL   [HTTP/1.1]
SERVER_SOFTWARE   [go]
```

When a client makes a POST request, such as with the following command

```
wget -O - -q --post-data="city=San%20Francisco" http://192.168.1.2:8080/show/weekly?mode=summary
```

the response looks the same except for the following line:

```
POST_DATA         [city=San%20Francisco]
```

##Fossil Example
The fossil distributed software management tool is a native executable that
uses a single SQLite database for all of its storage. It uses CGI for one of
its access methods. To set it up, use a cgi directive something like this in
your Caddyfile:

```
cgi /cgi-bin/* {match}
```

In your cgi-bin directory, make a file named, say, repo with the following contents:

```
#!/usr/bin/fossil
repository: /home/fossil/repo.fossil
```

Change the shebang line to reflect the location of the fossil executable, and the second line to reflect the
location of your fossil repository.


