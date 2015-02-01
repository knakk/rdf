ifndef BENCH
	BENCH := .
endif

bench_uncommited:
	@git stash && go test -run=NONE -bench=$(BENCH) > old.txt
	@git stash pop && go test -run=NONE -bench=$(BENCH) > new.txt
	benchcmp -mag=true old.txt new.txt
	@rm old.txt new.txt

bench_commited:
	@git checkout HEAD^ && go test -run=NONE -bench=$(BENCH) > old.txt
	@git checkout master && go test -run=NONE -bench=$(BENCH) > new.txt
	benchcmp -mag=true old.txt new.txt
	@rm old.txt new.txt