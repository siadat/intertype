test:
	@go get ./...
	@go run ./intertype/ ./testfiles/... 2> /tmp/got || true
	@cat /tmp/got | perl -pe 's#.*/(testfiles/.*)#\1#' > /tmp/got-relative
	@diff expected.txt /tmp/got-relative

vimdiff: test
	@vimdiff expected.txt /tmp/got

debug:
	@go run ./intertype/ -d ./testfiles/... 2> /tmp/got || true
