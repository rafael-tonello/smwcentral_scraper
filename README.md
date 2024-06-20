# About this project

this is a web crawler / scraper I write In Go to find Super Mario Wolrd hacks form the site [SMW Central](https://www.smwcentral.net/).

## How to use

To user it, just run it with the VSCode. You can, also, run it in the terminal with the command `go run main.go` or build it with the command `go build main.go` and run the generated file.

# about  PKV.SO and PKV package

PKV.so is a shared object written in C++ to work with a prefixtree. The PKV package helps to use this library in GO. In this project, the prefix tree is used as a database for internal operations of the crawler.

# .vscode folder
I use the VSCode to write the code, so I left a basic configuration along with the project

# special thanks
I want to send a especial thanks to the Flips team, that made the Flips patcher, and to the SMW Central team, that made the site and the hacks. Without them, this project would not be possible.

I also want to thank the 7zip team, that made the 7z library, that I use to extract the files and is a crucial part of the project.

FLips and 7z are not part of this project, but they are essential to it and are stored in the 'tools' folder.

# where are the games?
when you run the project, a folder name 'run' will be created in the root of the project. Inside it, you will find the games (in a folder named 'games').