##prettydiff

This app can be used to get pretty `git diff`. Opens an html version in your default browser.

This is a Go port of https://github.com/scottgonzalez/pretty-diff.

##Usage:

$ prettydiff

$ prettydiff 53aa6b98860f8e2a610d003a63b586e67396b003

give a commit hash to get it's diff.

#Getting the source

$ go get github.com/thewhitetulip/prettydiff

I recommend that you install it in $GOBIN by using the go install, so you can use prettydiff in any git repository.

###Dependencies:

We use http://github.com/skratchdot/open-golang/open to launch the default browser.

License: MIT.

##Screenshot

![Home Page](screenshot.png)


Made with :heart: in India.
