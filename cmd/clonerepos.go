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
	log "github.com/sirupsen/logrus"
)

var (
	Cmd             string
	ReposList       string
	CloneDir        string
	AddSshRemote    bool
	SshUserName     string
	SshRemoteName   string
	CredentialsUser string
	CredentialsPass string
)

type Repo struct {
	Url string
}

func parseArgs() {

	cloneFlagSet := flag.NewFlagSet("clone", flag.ExitOnError)

	cloneFlagSet.StringVar(&ReposList, "in", "", "Input file for the repos")
	cloneFlagSet.StringVar(&CloneDir, "out", "", "Output directory for the repos")
	cloneFlagSet.BoolVar(&AddSshRemote, "add-ssh-remote", false, "Add ssh remote along with http(s)")
	cloneFlagSet.StringVar(&SshUserName, "ssh-user", "git", "Ssh user to access the repo")
	cloneFlagSet.StringVar(&SshRemoteName, "ssh-remote-name", "", "Remote name for ssh access.")
	cloneFlagSet.StringVar(&CredentialsUser, "user", "", "Auth: repo user")
	cloneFlagSet.StringVar(&CredentialsPass, "pass", "", "Auth: repo password")

	// Verify that a sub-command has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the sub-command
	if len(os.Args) < 2 {
		log.Warn("sub command is required. e.g. clone")
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
			log.Error("You did not supplied input file name")
			os.Exit(1)
		}

		if len(CloneDir) == 0 {
			log.Error("You did not supplied output directory")
			os.Exit(1)
		}
		if AddSshRemote {
			if SshUserName == "" {
				log.Error("Please specify username for ssh access.")
				os.Exit(1)
			}
			if SshRemoteName == "" {
				log.Error("Please specify name for ssh remote. e.g. 'github', 'upstream', etc...")
				os.Exit(1)
			}
		}
		log.Infof("Using %s file for incoming repos\n", ReposList)
		log.Infof("Using %s output directory for the repos\n", CloneDir)
		Cmd = "clone"
	}

}

func ReadRepos(filename string) (*[]Repo, int) {

	var res []Repo

	file, err := os.Open(filename)

	if err != nil {
		log.Errorf("Error opening %s\n", filename)
		return nil, 0
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("Error reading line file: %s \n", err)
			os.Exit(1)
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		repo := Repo{
			Url: strings.TrimSpace(line),
		}

		res = append(res, repo)
	}

	return &res, len(res)
}

// Info should be used to describe the example commands that are about to run.
//func Info(format string, args ...interface{}) {
//	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
//}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	// TODO - refactorin
	log.Errorf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func CloneCmd(repos *[]Repo) error {
	for _, r := range *repos {
		log.Infof("git clone %s", r.Url)
		cloneDirName, err := getRepoNameFromUrl(r.Url)
		if err != nil {
			log.Debugf("Skip %s", r.Url)
			continue
		}
		log.Info("CloneDirName = ", cloneDirName)
		cloneDir := CloneDir + cloneDirName
		log.Info(fmt.Sprintf("Cloning in %s\n", cloneDir))

		exist, err := checkIfRepoAlreadyExist(cloneDir)
		if err != nil {
			log.Errorf("Can not clone repo %s, %w\n", r.Url, err)
		}
		if !exist {
			repo, err := cloneRepo(r.Url, cloneDir)
			CheckIfError(err)
			if AddSshRemote {
				addRemoteToRepo(repo, r.Url, SshUserName, SshRemoteName, true)
			}
		} else {
			log.Infof("Repo %s is already cloned\n", r.Url)
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

func addRemoteToRepoInDir(repoDir, httpUrl, sshUser, remoteName string, replaceIfExist bool) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}
	err = addRemoteToRepo(repo, httpUrl, sshUser, SshRemoteName, replaceIfExist)
	return err
}

func generateSSHRemoteUrl(httpUrl, sshUser string) string {
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
			log.Errorf("Error deleting remote %s \n", remoteName)
		}
	}

	sshUrl := generateSSHRemoteUrl(httpUrl, sshUser)
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

	log.Info("Cloned...")
	return r, nil
}

// TODO:
// Add clone command - -in <repo_list> -out <store_dir> -u <user> -p <pass>/<token> // support https
// Add add_ssh_remote -repo_dir <repo_dir> -remote_name <remote_name> -type <ssh> -ssh_user <git> - add remote - e.g. - github git@github:/reponame.git
// Add fetch_all - repo_dir - iterate and fetch_all

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	parseArgs()
	repos, size := ReadRepos(ReposList)

	if size == 0 {
		log.Warn("No repos defined for cloning")
		os.Exit(1)
	}

	log.Info("----------------------------------------------------")

	if Cmd == "clone" {
		CloneCmd(repos)
	}

}
