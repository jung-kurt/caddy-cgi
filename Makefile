all : README.md root/index.html doc.go

root/index.html : doc/hdr.txt doc/mid2.txt doc/ftr.txt
	cat $^ > $@
	tidy -quiet -output /tmp/0 $@

doc/mid1.txt : doc/index.md
	sed \
	-e '/^\[!\[/d' \
	$< | markdown > $@

doc/mid2.txt : doc/mid1.txt
	tr '\n' '\001' < $< | sed \
	-e 's/<\/div>/\f/g' \
	-e 's/<\/p>/\v/g' \
	-e 's/class="tag"/class="tag addon"/g' \
	-e 's/<div class="block"><p>\([^\f]*\)\v\f/<p><mark class="block">\1<\/mark>\v/g' \
	-e 's/<div class="code"><p>\([^\v]*\)\v\f/<p><code class="block">\1<\/code>\v/g' \
	-e 's/\v/<\/p>/g' \
	-e 's/\f/<\/div>/g' \
	| tr '\001' '\n' > $@

README.md : doc/index.md
	sed \
	-e '/^> %addon-message%/,/github\.com/d' \
	-e 's/# cgi.*/# cgi/g' \
	-e '/^> %block%/d' \
	-e 's/^###/##/g' \
	-e '/^> %code%/,/^[^>]/s/^> /\t/g' \
	-e '/\t%code%/d' \
	-e 's/^> //g' \
	-e '/\]: class:/d' \
	-e 's/\[\([^]]*\)\]\[arg\]/\1/g' \
	-e 's/\[\([^]]*\)\]\[dir\]/\1/g' \
	-e 's/\[\([^]]*\)\]\[subdir\]/\1/g' \
	-e 's/\*\.lua/\x01/g' \
	-e 's/\*//g' \
	-e 's/\x01/*.lua/g' \
	$^ > $@

doc.go : README.md
	printf "/*\n" > $@
	sed \
	-e 's/`//g' \
	-e 's/^##* //g' \
	-e '/^cgi$$/d' \
	-e 's/\*//g' \
	-e '/\[.*\]:/d' \
	-e 's/\[\([^]]*\)\]\[[^]]*\]/\1/g' \
	-e '/^!\[/d' \
	-e '/./,$$!d' \
	$^ >> $@
	printf "\n*/\npackage cgi\n" >> $@

clean :
	rm -f doc/mid*txt README.md doc.go root/index.html
