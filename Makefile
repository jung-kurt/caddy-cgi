all : ok documentation lint

documentation : doc/index.html doc.go README.md 

lint : ok/index.html

ok : 
	mkdir ok

ok/%.html : doc/%.html
	tidy -quiet -output /dev/null $<
	touch $@

cov : all
	go test -v -coverprofile=coverage && go tool cover -html=coverage -o=coverage.html

check :
	golint .
	go vet -all .
	gofmt -s -l .
	goreportcard-cli -v

README.md : doc/document.md
	pandoc --read=markdown --write=gfm < $< > $@

doc/index.html : doc/document.md doc/html.txt doc/caddy.xml
	pandoc --read=markdown --write=html --template=doc/html.txt \
		--metadata pagetitle="CGI for Caddy" --syntax-definition=doc/caddy.xml < $< > $@

doc.go : doc/document.md doc/go.awk
	pandoc --read=markdown --write=plain $< | awk --assign=package_name=cgi --file=doc/go.awk > $@
	gofmt -s -w $@

build :
	cd ../caddy-custom
	go build -v
	sudo setcap cap_net_bind_service=+ep ./caddy
	./caddy -plugins | grep cgi
	./caddy -version

clean :
	rm -f coverage.html coverage ok/* doc/index.html doc.go README.md
