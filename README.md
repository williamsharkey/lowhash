# _lowhash_
Join the search for magical  strings which produce the lowest sha256 hashes.

## Get on the Scoreboard

See the current 256 lowest strings at https://lowhash.com

## Install

  Download a compiled executable for your system:
  
  - Windows [64bit](https://lowhash.com/build/win64/lowhash.exe) [32bit](https://lowhash.com/build/win32/lowhash.exe)  
  - macOS   [64bit](https://lowhash.com/build/osx64/lowhash) [32bit](https://lowhash.com/build/osx32/lowhash)
  - Linux   [64bit](https://lowhash.com/build/linux64/lowhash) [32bit](https://lowhash.com/build/linux32/lowhash)
  
 ##Build from Source
  
  If you want to build it yourself:
  
  - you need the Go compiler, see https://golang.org
  - run: `go get github.com/williamsharkey/lowhash`
  - change directory to $GOPATH/scr/github.com/williamsharkey/lowhash/
  - `go build`

## About
  - lowhash loads a default sentence generator grammar unless grammar.txt is found in the current directory
  - run `lowhash` and wait for it to produce strings with low hashes
  - low hashes which are found will be automatically posted to https://lowhash.com
  - low hashes will be printed to console, so if your internet is inaccessable, so look at the console
  
 
## tag yourself!
  - lowhash takes two optional arguments for specifying prefix and postfix.
  
### examples
```
  lowhash                           => "The quick brown fox."
  lowhash "" ""                     => "The quick brown fox."
  lowhash "(william) "              => "(william) The quick brown fox."
  lowhash "<yo> " " -william"       => "<yo> The quick brown fox -william"
```


## note

feel free to build your own faster search program. If you find a good string you can still post it to https://lowhash.com, programatically via a post request or manually by pasting it into the input box at the website.

https://lowhash.com is currently interested in collecting strings which are relatively short, 256 characters or less.
