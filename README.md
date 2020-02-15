# lowhash
Search for (readable) strings which produce low sha256 hashes.

# installation

  If you have Windows, you can download the executable from the "releases"
  
  If you want to build it yourself, or are on OSX or Linux
  
  - make sure you have go installed, see [https://golang.org]
  - `go get github.com/williamsharkey/lowhash`
  - change directory to <your go path>/scr/github.com/williamsharkey/lowhash/
  - `go build`
  - now you should have produced lowhash.exe(windows) or lowhash

## run
  - cd to folder with lowhash
  - run `lowhash` and wait for it to produce strings with low hashes
  - low hashes which are found will be automatically posted to https://lowhash.com
  - low hashes will be printed to console, so if your internet is inaccessable, so look at the console
  
 
## tag yourself!
  - lowhash takes two optional arguments for specifying prefix and postfix.
  
### examples
```
  lowhash                    (no prefix / default period postfix)   => "The quick brown fox."
  lowhash "" ""              (no prefix / no postfix    )           => "The quick brown fox"
  lowhash "(william)"                                               => "(william) The quick brown fox."
  lowhash "" " (william)"                                           => "The quick brown fox (william)"
```


## note

feel free to build your own faster search program. If you find a good string you can still post it to https://lowhash.com, programatically via a post request or manually by pasting it into the input box at the website.

https://lowhash.com is currently interested in collecting strings which are relatively short, 256 characters or less.
