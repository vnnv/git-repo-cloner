run:
	go run main.go

clonerepos:
	go run cmd/clonerepos.go clone -in local/repos.txt -out /mnt/devel/projects/personal/go-by-example/git-cloner/local/repos -add-ssh-remote=true -ssh-user=git -ssh-remote-name=github

test:
	go test -cover -race ./...