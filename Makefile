test:
	mkdir -p tmp
	go test -v -count=1 -race ./...
	rm -rf tmp

doc:
	godoc --http=:6060