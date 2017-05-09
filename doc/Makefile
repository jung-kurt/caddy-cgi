all: index.html ../README.md ../doc.go

body.html : cgi.md
	markdown -f +links,+image,+smarty,+ext,+divquote -o body.html $^

index.html : hdr.html body.html ftr.html
	cat $^ > $@
	tidy -q -o /dev/null $@

cgi.md ../README.md ../doc.go : doc.txt
	lua doc.lua

clean :
	rm -f cgi.md ../README.md ../doc.go

