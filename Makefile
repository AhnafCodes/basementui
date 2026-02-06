.PHONY: build test demo example1 example2 example3 example4 example5 example6 example7 example8 example9 example10 example11 example12 example12-chroma

build:
	cd go && go build ./...

test:
	cd go && go test ./...

demo:
	cd go && go run cmd/demo/main.go

example1:
	cd go && go run cmd/example1_hello/main.go

example2:
	cd go && go run cmd/example2_counter/main.go

example3:
	cd go && go run cmd/example3_computed/main.go

example4:
	cd go && go run cmd/example4_clock/main.go

example5:
	cd go && go run cmd/example5_progress/main.go

example6:
	cd go && go run cmd/example6_conditional/main.go

example7:
	cd go && go run cmd/example7_input/main.go

example8:
	cd go && go run cmd/example8_textinput/main.go

example9:
	cd go && go run cmd/example9_list/main.go

example10:
	cd go && go run cmd/example10_layout/main.go

example11:
	cd go && go run cmd/example11_markdown/main.go

example12:
	cd go && go run cmd/example12_chroma/main.go

example12-chroma:
	cd go && go run -tags chroma cmd/example12_chroma/main.go
