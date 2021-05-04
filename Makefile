.PHONY: test
test: test_runner
	bin/test_runner transforms/t1/t1-test.star

test_runner:
	mkdir -p bin/
	go build -o bin/test_runner

clean:
	rm -rf bin/
