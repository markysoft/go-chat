watch:
	templ generate --watch --proxy="http://localhost:3000" --cmd="go run ." --open-browser=false

test:
	go test ./...

fmt:
	go fmt ./...
