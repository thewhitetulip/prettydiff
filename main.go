package main

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

func main() {
	var index []int
	var cmd *exec.Cmd

	if len(os.Args) > 1 {
		diffID := os.Args[1]
		cmd = exec.Command("git", "diff", diffID)
	}

	cmd = exec.Command("git", "diff")

	// open the out file for writing
	outfile, err := os.Create("/tmp/diff.txt")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		if err.Error() == "exit status 129" {
			fmt.Println("Not a git repository")
			os.Exit(1)
		}
	}
	cmd.Wait()

	file, err := ioutil.ReadFile("/tmp/diff.txt")
	if err != nil {
		fmt.Println("Please check if you have access to /tmp folder")
		os.Exit(1)
	}
	htmlFile, err := os.OpenFile("/tmp/diff.html", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Println("Please check if you have access to /tmp folder")
	}
	strFile := string(file)
	lines := strings.Split(strFile, "\n")

	if strFile == "" {
		fmt.Println("No diff to show")
		os.Exit(0)
	}

	if len(lines) > 0 {
		io.WriteString(htmlFile, basehtml)
	}

	for i := 0; i < len(lines); i++ {
		a := strings.Split(lines[i], " ")[0]
		if a == "diff" {
			index = append(index, i)
		}
	}
	var lower, higher int
	for i := 0; i < len(index); i++ {
		lower = index[i]

		if i == len(index)-1 {
			higher = len(lines) - 1
		} else {
			higher = index[i+1]
		}
		analyzeLines(lines[lower:higher], htmlFile, lower)
	}

	htmlFile.WriteString("</div></div></body>")

	open.Start("/tmp/diff.html")
}

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
