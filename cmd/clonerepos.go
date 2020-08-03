package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

var (
	ReposList     string
	CloneDir      string
	Cmd           string
	AddSshRemote  bool
	SshUserName   string
	SshRemoteName string
)

type Repo struct {
	Url string
}

func parseArgs() {

	cloneFlagSet := flag.NewFlagSet("clone", flag.ExitOnError)

	cloneFlagSet.StringVar(&ReposList, "in", "", "Input file for the repos")
	cloneFlagSet.StringVar(&Cmd, "cmd", "clone", "git command. Available commands - clone, list")
	cloneFlagSet.StringVar(&CloneDir, "out", "", "Output directory for the repos")
	cloneFlagSet.BoolVar(&AddSshRemote, "add-ssh-remote", false, "Add ssh remote along with http(s)")
	cloneFlagSet.StringVar(&SshUserName, "ssh-user", "git", "Ssh user to access the repo")
	cloneFlagSet.StringVar(&SshRemoteName, "ssh-remote-name", "", "Remote name for ssh access.")

	// Verify that a sub-command has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the sub-command
	if len(os.Args) < 2 {
		fmt.Println("sub command is required. e.g. clone")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "clone":
		cloneFlagSet.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if cloneFlagSet.Parsed() {
		if len(ReposList) == 0 {
			fmt.Println("You did not supplied input file name")
			os.Exit(1)
		}

		if len(CloneDir) == 0 {
			fmt.Println("You did not supplied output directory")
			os.Exit(1)
		}
		if AddSshRemote {
			if SshUserName == "" {
				fmt.Println("Please specify username for ssh access.")
				os.Exit(1)
			}
			if SshRemoteName == "" {
				fmt.Println("Please specify name for ssh remote. e.g. 'github', 'upstream', etc...")
				os.Exit(1)
			}
		}
		fmt.Printf("Using %s file for incoming repos\n", ReposList)
		fmt.Printf("Using %s output directory for the repos\n", CloneDir)
		fmt.Printf("Command is %s\n", Cmd)

	}

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

func CloneCmd(repos *[]Repo) error {
	for _, r := range *repos {
		Info("git clone " + r.Url)
		cloneDirName, err := getRepoNameFromUrl(r.Url)
		if err != nil {
			fmt.Println("Skip ", r.Url)
			continue
		}
		fmt.Println("CloneDirName = ", cloneDirName)
		cloneDir := CloneDir + cloneDirName
		Info(fmt.Sprintf("Cloning in %s\n", cloneDir))

		exist, err := checkIfRepoAlreadyExist(cloneDir)
		if err != nil {
			fmt.Printf("Can not clone repo %s, %w\n", r.Url, err)
		}
		if !exist {
			repo, err := cloneRepo(r.Url, cloneDir)
			CheckIfError(err)
			if AddSshRemote {
				addRemoteToRepo(repo, r.Url, SshUserName, SshRemoteName, true)
			}
 		}else{
			fmt.Printf("Repo %s is already cloned\n", r.Url)
			if AddSshRemote {
				addRemoteToRepoInDir(cloneDir, r.Url, SshUserName, SshRemoteName, true)
			}
			// Update and if needed add ssh origin
		}
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

func checkIfRepoAlreadyExist(repoDir string) (bool, error) {
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return false, nil // no target dir
	}
	// check if directory is empty
	targetDir, err := os.Open(repoDir)
	if err != nil {
		return false, err
	}
	defer targetDir.Close()

	_, err = targetDir.Readdirnames(1)
	if err == io.EOF {
		return false, nil // target dir exist, but is empty
	}

	// Try to open the repo
	_, err = git.PlainOpen(repoDir)
	if err != nil {
		// Can not open the repo - seems the directory is not empty and is not repo
		return false, err
	}
	// Successfully check that the target dir is repo and can be opened
	return true, nil
}

func addRemoteToRepoInDir(repoDir, httpUrl, sshUser, remoteName string, replaceIfExist bool ) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}
	err = addRemoteToRepo(repo, httpUrl, sshUser, SshRemoteName, replaceIfExist)
	return err
}

func generateSshRemoteUrl(httpUrl, sshUser string) string {
	var tempUrl = httpUrl
	if strings.HasPrefix(httpUrl, "ftp://") {
		tempUrl = httpUrl[6:]
	}
	if strings.HasPrefix(httpUrl, "http://") {
		tempUrl = httpUrl[7:]
	}
	if strings.HasPrefix(httpUrl, "https://") {
		tempUrl = httpUrl[8:]
	}
	tempUrl = strings.Replace(tempUrl, "/", ":", 1)
	return sshUser + "@" + tempUrl
}

func addRemoteToRepo(repo *git.Repository, httpUrl, sshUser, remoteName string, replaceIfExist bool) error {
	if repo == nil {
		return nil
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}
	remoteAlreadyExist := false
	for _, r := range remotes {
		if r.String() == remoteName {
			// remote already exist
			remoteAlreadyExist = true
		}
 	}
 	if replaceIfExist && remoteAlreadyExist {
 		// delete old remote
		err := repo.DeleteRemote(remoteName)
		if err != nil {
			fmt.Printf("Error deleting remote %s \n", remoteName)
		}
	}

	sshUrl := generateSshRemoteUrl(httpUrl, sshUser)
	// Create new remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{sshUrl},
	})

	return err
}

func cloneRepo(repoUrl, cloneDir string) (*git.Repository, error) {
	r, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:               repoUrl,
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
	return r, nil
}

// TODO:
// Add clone command - -in <repo_list> -out <store_dir> -u <user> -p <pass>/<token> // support https
// Add add_ssh_remote -repo_dir <repo_dir> -remote_name <remote_name> -type <ssh> -ssh_user <git> - add remote - e.g. - github git@github:/reponame.git
// Add fetch_all - repo_dir - iterate and fetch_all

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
