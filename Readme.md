Git repo cloner tool
===

See Makefile

---
Available options:
./executable clone
 
    * `-in` - input file containing a repository list - every line is a single repo url (http/https)
    * `-out` - target directory where to put the repos
    * `-add-ssh-remote` true/false - if true - to every cloned repo add new ssh remote
    * `-ssh-user` - username used to access the repo via ssh
    * `-ssh-remote-name` - name of the remote   

example:

```bash
$ ./git-cloner clone -in input_file.csv -out /target_dir -add-ssh-remote=true -ssh-user=git -ssh-remote-name=gitgub
```

TODO:
    
    * Add username and token for github authentication
    * Add Gitlab support
    * Add option to query github via rest interface
    
    