package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

var (
	ReposList string
	CloneDir  string
	Cmd       string
)

type Repo struct {
	Url string
}

func parseArgs() {
	flag.StringVar(&ReposList, "in", "", "Input file for the repos")
	flag.StringVar(&Cmd, "cmd", "clone", "git command. Available cmds - clone, list")
	flag.StringVar(&CloneDir, "out", "", "Output directory for the repos")

	flag.Parse()

	if len(ReposList) == 0 {
		fmt.Println("You did not supplied input file name")
		os.Exit(1)
	}

	if len(CloneDir) == 0 {
		fmt.Println("You did not supplied output directory")
		os.Exit(1)
	}
	fmt.Printf("Using %s file for incoming repos\n", ReposList)
	fmt.Printf("Using %s output directory for the repos\n", CloneDir)
	fmt.Printf("Command is %s\n", Cmd)

}

func ReadRepos(filename string) (*[]Repo, int) {

	var res []Repo

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		fmt.Printf("Error opening %s\n", filename)
		os.Exit(1)
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Errorf("Error reading line file: %s \n", err)
			break
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		//fmt.Printf("%s", line)

		repo := Repo{
			Url: strings.TrimSpace(line),
		}

		res = append(res, repo)
	}

	return &res, len(res)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func CloneCmd(repos *[]Repo) (error) {
	for _, r := range *repos {
		Info("git clone " + r.Url)
		cloneDirName, e := getRepoNameFromUrl(r.Url)
		if e != nil {
			fmt.Println("Skip ", r.Url)
			continue
		}
		fmt.Println("CloneDirName = ", cloneDirName)
		cloneDir := CloneDir + cloneDirName
		Info(fmt.Sprintf("Cloning in %s\n", cloneDir))
		err := cloneRepo(r.Url, cloneDir)
		CheckIfError(err)
	}
	return nil
}

func getRepoNameFromUrl(repo string) (string, error) {
	if len(repo) == 0 {
		return "", fmt.Errorf("Invalid repo url\n")
	}
	lastSlash := strings.LastIndex(repo, "/")
	extIndex := strings.LastIndex(repo, ".git")
	return repo[lastSlash:extIndex], nil
}

func cloneRepo(repourl, cloneDir string) (error){
	_, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:               repourl,
		Auth:              nil,
		RemoteName:        "",
		ReferenceName:     "",
		SingleBranch:      false,
		NoCheckout:        false,
		Depth:             0,
		RecurseSubmodules: 0,
		Progress:          nil,
		Tags:              0,
	})
	CheckIfError(err)

	Info("Cloned...")
	return nil
}
// TODO:
// Add clone command - -in <repo_list> -out <store_dir> -u <user> -p <pass>/<token> // support https
// Add add_ssh_remote -repo_dir <repo_dir> -remote_name <remote_name> -type <ssh> -ssh_user <git> - add remote - e.g. - github git@github:/reponame.git
// Add fetch_all -repo_dir - iterate and fetch_all

func main() {
	parseArgs()
	repos, size := ReadRepos(ReposList)

	if size == 0 {
		fmt.Println("No repos defined for cloning")
		os.Exit(1)
	}

	fmt.Println("----------------------------------------------------")

	if Cmd == "clone" {
		CloneCmd(repos)
	}


}
