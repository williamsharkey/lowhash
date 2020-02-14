# lowhash
search for (readable) strings which produce low sha256 hashes.

# installation

  If you have Windows, you can download the executable from the "releases"
  
  If you want to build it yourself, or are on OSX or Linux
  
  - make sure you have the go installed, then:
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
  - If you want 'credit' for your finds, set a prefix or postfix for your search strings.   

  - if you provide arguments to lowhash, they will be taken as the prefix and postfix.
    - a default postfix of period is added, which can be customized
  
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
