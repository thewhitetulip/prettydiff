package main

/*
This program runs git diff, gets the output and creates a coloured html page in the location /tmp/diff.html and opens it in your browser.
Will not work with Windows since it doesn't have a /tmp directory, but might work if you have cygwin installed, which probably has it's own /tmp directory. Please send a PR if you figure out how to make this work in Windows.
*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const basehtml = `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Pretty Diff</title>
    <style> 
        #wrapper {
            display: inline-block;
            margin-top: 1em;
            min-width: 800px;
            text-align: left;
        }
        
        h2 {
            background: #fafafa;
            background: -moz-linear-gradient(#fafafa, #eaeaea);
            background: -webkit-linear-gradient(#fafafa, #eaeaea);
            -ms-filter: "progid:DXImageTransform.Microsoft.gradient(startColorstr='#fafafa',endColorstr='#eaeaea')";
            border: 1px solid #d8d8d8;
            border-bottom: 0;
            color: #555;
            font: 14px sans-serif;
            overflow: hidden;
            padding: 10px 6px;
            text-shadow: 0 1px 0 white;
            margin: 0;
        }
        
        .file-diff {
            border: 1px solid #d8d8d8;
            margin-bottom: 1em;
            overflow: auto;
            padding: 0.5em 0;
        }
        
        .file-diff > div {
            width: 100%:
        }
        
        pre {
            margin: 0;
            font-family: "Bitstream Vera Sans Mono", Courier, monospace;
            font-size: 12px;
            line-height: 1.4em;
            text-indent: 0.5em;
        }
        
        .file {
            color: #aaa;
        }
        
        .delete {
            background-color: #fdd;
        }
        
        .insert {
            background-color: #dfd;
        }
        
        .info {
            color: #a0b;
        }
    </style>
</head>
<body>
<div id="wrapper">`

const (
	helpMessage           = "prettyprint:\nWill print the git diff output in a visual way in your browser.\nIf no parameter is provided, it shows output of git diff, otherwise it shows the git diff at that commit ID.\nUsage: \n\tprettydiff\n\tprettydiff 53aa6b98860f8e2a610d003a63b586e67396b003"
	tmpDirErrMsg          = "Do not have access to the /tmp directory"
	noDiffsErrMsg         = "No diff to show"
	noGitRepoErrMsg       = "Not a git repository"
	invalidCommitIDErrMsg = "Please provide valid commit ID"
	tmpFilePath           = "/tmp/diff.txt"
	tmpHTMLPath           = "/tmp/diff.html"
)

func main() {
	var index []int //stores the line number of each diff block
	var cmd *exec.Cmd

	switch len(os.Args) {
	//the first commandline arg is the program name
	case 1:
		cmd = exec.Command("git", "diff")
	//only two arguments are supported, -h and the commit ID
	case 2:
		diffArg := os.Args[1]

		if diffArg == "-h" {
			fmt.Println(helpMessage)
			os.Exit(0)
		} else {
			//length of valid commit ID is 40,
			if len(diffArg) == 40 {
				cmd = exec.Command("git", "diff", diffArg)
			} else {
				fmt.Println(invalidCommitIDErrMsg)
				os.Exit(1)
			}
		}
	}

	// open the out file for writing
	outfile, err := os.Create(tmpFilePath)
	if err != nil {
		fmt.Println(tmpDirErrMsg)
		os.Exit(1)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		if err.Error() == "exit status 129" {
			fmt.Println(noGitRepoErrMsg)
			os.Exit(1)
		}
	}
	cmd.Wait()

	file, err := ioutil.ReadFile(tmpFilePath)
	if err != nil {
		fmt.Println(tmpDirErrMsg)
		os.Exit(1)
	}
	htmlFile, err := os.OpenFile(tmpHTMLPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Println(tmpDirErrMsg)
	}
	strFile := string(file)
	lines := strings.Split(strFile, "\n")

	//will be displayed when prettydiff is run on git repository where no diffs are present
	if strFile == "" {
		fmt.Println(noDiffsErrMsg)
		os.Exit(0)
	}

	//write the base to the html page, the static part.
	if len(lines) > 0 {
		io.WriteString(htmlFile, basehtml)
	}

	//find out the indices where the new git block start
	for key, line := range lines {
		if strings.HasPrefix(line, "diff") {
			index = append(index, key)
		}
	}

	var lower, higher int
	for key, value := range index {
		lower = value
		//The last index is from last value of index to the last line in the diff.txt file
		if key == len(index)-1 {
			higher = len(lines) - 1
		} else {
			higher = index[key+1]
		}
		analyzeLines(lines[lower:higher], htmlFile, lower)
	}

	htmlFile.WriteString("</div></div></body>")

	open.Start(tmpHTMLPath)
}

//analyzeLines takes a slice of a section of a diff and creates one block
func analyzeLines(lines []string, htmlFile *os.File, lower int) {
	for _, line := range lines {
		//escape HTML if any
		var name string
		line = strings.Replace(line, "<", "&lt;", -1)
		line = strings.Replace(line, ">", "&gt;", -1)
		line = strings.Replace(line, "&", "&amp;", -1)

		if strings.HasPrefix(line, "-") {
			line = "\n<pre class='delete'>" + line + "</pre>"
		} else if strings.HasPrefix(line, "+") {
			line = "\n<pre class='insert'>" + line + "</pre>"
		} else if strings.HasPrefix(line, "diff") {
			a := strings.Split(line, " ")
			name = a[2][2:]

			line = "\n<pre class='info'>" + line + "</pre>"
			if lower != 0 {
				htmlFile.WriteString("</div></div>\n")
			}

			htmlFile.WriteString("\n<h2>" + name + "</h2>")
			htmlFile.WriteString("\n<div class='file-diff'><div>")
			line = "\n<pre class='file'>" + line + "</pre>"
		} else if strings.HasPrefix(line, "@") {
			line = "\n<pre class='file'>" + line + "</pre>"
		} else if strings.HasPrefix(line, "index") {
			line = "\n<pre class='info'>" + line + "</pre>"
		} else {
			line = "\n<pre class='context'>" + line + "</pre>"
		}

		htmlFile.WriteString(line)
	}
}
