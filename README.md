Small utility to import local git repos into a gitea server
===========================================================

*Must be run on the gitea server*

The folder structure should be in the format of organization/repo 

If the organization is missing in GITEA it will be created on the fly.

Example:
```
 + ToMigrate
      |
	  + OrgA
         |
		 + RepoX.git
         + RepoY.git
      + OrgB
         |
         + RepoZ.git
```
Command-line syntax:
```
Usage of ./gitea-migrate:
  -debug
    	Show debug output
  -password string
    	GITEA password
  -path string
    	path to local repos
  -url string
    	URL to GITEA. Ex: https://git.server.com
  -username string
    	GITEA username
```

Example:
```
./gitea-migrate --path /mnt/repos/git --url https://git.server.com --username someuser --password xxxxxxx
```
