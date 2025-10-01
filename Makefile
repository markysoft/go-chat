watch:
	templ generate --watch --proxy="http://localhost:3000" --cmd="go run ." --open-browser=false

# Generate templ files once for debugging
gen:
	templ generate

# Watch templ files without running the server (for debugging)
watch-templ:
	templ generate --watch

# Debug mode - generate once then run with delve
debug: gen
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

test:
	go test ./...

fmt:
	go fmt ./...
