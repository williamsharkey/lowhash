package main

import (
	nativeSha "crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"strings"
	"time"
)

func main() {

	prefix := ""
	postfix := "."

	if len(os.Args) > 1 {
		prefix = os.Args[1]
	}
	if len(os.Args) > 2 {
		postfix = os.Args[2]
	}

	fmt.Println(" ")
	fmt.Println(" +-----------+")
	fmt.Println(" |  lowhash  |")
	fmt.Println(" +-----------+")
	fmt.Println()
	fmt.Println(" https://lowhash.com")
	fmt.Println()
	fmt.Println()
	fmt.Printf("prefix: %q\n", prefix)
	fmt.Printf("postfix: %q\n", postfix)
	fmt.Println("to set a prefix/postfix, run with arguments, eg: lowhash \"my prefix\" \"my postfix\"")
	fmt.Println()
	start := time.Now()

	// prime the program with a cutoff
	atleast, err := post("a new paper suffers as if the thoughts", []byte{0xff})

	const batch = 400000 * 2 // about every two seconds, modulo how fast your computer is

	initWordsByte()
	//sx,n := randSentenceFast([]byte(prefix),[]byte(postfix))
	//fmt.Println(string(sx[:n]))
	//return
	for i := 0; true; i++ {

		//s := randSentence(prefix, postfix)
		s, n := randSentenceFast([]byte(prefix), []byte(postfix))
		//s:="fixed"
		hash := hashFast(s[:n])
		//hash := hash(s)
		if leftLess(hash, atleast) {
			atleast, err = post(string(s[:n]), atleast)
			if err != nil {
				fmt.Println("Quitting, err " + err.Error())
				return
			}
		}

		if i > batch {

			hashesPerSecond := float64(batch) / (float64(time.Since(start).Nanoseconds()) / float64(1000000000))
			start = time.Now()
			fmt.Printf("%dk hashes per second \n", int64(hashesPerSecond/1000))

			// avoid an int overflow
			i = 0
		}
	}
}

func leftLess(a, b []byte) bool {
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false

		}
	}
	return false
}

var cutoffRegex, _ = regexp.Compile(".*cutoff: (.*)")
var msgRegex, _ = regexp.Compile(".*msg: (.*)")
var rankRegex, _ = regexp.Compile(".*rank: (.*)")

func post(s string, atleast []byte) (atleastOut []byte, fail error) {
	atleastOut = atleast
	form := url.Values{}
	form.Add("msg", s)
	sageURL := "https://lowhash.com"
	data := url.Values{}
	data.Add("msg", s)
	u, errParse := url.ParseRequestURI(sageURL)
	if errParse != nil {
		fmt.Println("errparse not nil")
		return
	}
	u.Path = "/sagebird"
	urlStr := u.String()

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	if err != nil {
		fmt.Printf("fail req %s ", err.Error())
		return
	}
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		fmt.Printf("fail resp %s ", err.Error())

	} else {

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				fmt.Println("err")
				return
			}
			bodyString := string(bodyBytes)
			if strings.HasPrefix(bodyString, "\nhash does not rank") {
				a := cutoffRegex.FindStringSubmatch(bodyString)[1]
				atleastOut = stringToHexBytes(a)
				fmt.Println("Set cutoff to " + a[0:8] + " " + a[8:16])
			} else if strings.HasPrefix(bodyString, "\nrank: ") {
				b := cutoffRegex.FindStringSubmatch(bodyString)[1]
				atleastOut = stringToHexBytes(b)
				rank := rankRegex.FindStringSubmatch(bodyString)[1]
				msg := msgRegex.FindStringSubmatch(bodyString)[1]
				fmt.Println("Excellent! You placed " + rank + " with: \"" + msg + "\"")
				fmt.Println(" Update cutoff to " + b[0:8] + " " + b[8:16])
			} else if strings.HasPrefix(bodyString, "\nalready present") {
				c := cutoffRegex.FindStringSubmatch(bodyString)[1]
				atleastOut = stringToHexBytes(c)
				fmt.Println("Set cutoff to " + c[0:8] + " " + c[8:16])
			} else {
				fmt.Print("unknown response")
				fmt.Print(bodyString)
			}

		} else {
			fmt.Printf("status code %v\n", resp.StatusCode)
			fail = fmt.Errorf("%d", resp.StatusCode)

		}
	}
	return
}

func stringToHexBytes(s string) []byte {
	data, _ := hex.DecodeString(s)
	return data
}

func hash(x string) string {
	sha.Reset()

	io.WriteString(sha, x)
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func hashFast(x []byte) []byte {
	sha.Reset()
	sha.Write(x)
	return sha.Sum(nil)
	//return fmt.Sprintf("%x", sha.Sum(nil))
}

func randSentence(pre, post string) string {
	sentence := pre
	for i := 0; i < len(words); i++ {
		var j = len(words[i])
		var r = rand.Intn(j)
		if i > 0 {
			sentence += " "
		}
		sentence += words[i][r]
	}
	return sentence + post
}

func randSentenceFast(pre, post []byte) (sentence [256]byte, n int) {
	n = len(pre)
	copy(sentence[:], pre)

	//sentence := pre
	for i := 0; i < len(wordsByte); i++ {
		var j = len(wordsByte[i])
		var r = rand.Intn(j)
		if i > 0 {
			sentence[n] = []byte(" ")[0]
			n = n + 1
		}
		copy(sentence[n:], wordsByte[i][r])
		n += len(wordsByte[i][r])
	}

	copy(sentence[n:], post)
	n += len(post)
	return
	//return sentence, n
}

var sha = nativeSha.New()

func init() {
	rand.Seed(time.Now().Unix())

}

var words = [][]string{{
	"Capture", "Dispense", "Eject", "Imagine", "Be", "Do", "Have", "Become", "Once", "If", "Imagine if", "Suppose that",
	"I know", "Don't fear", "Because", "Therefore", "As though", "We think", "I believe", "You hope"},

	{"those", "my", "her", "his", "what", "our", "when", "that", "some", "one", "a", "some", "the", "a", "the"},

	{"lost", "hidden", "sorry", "embodied", "intrinsic", "electronic", "new", "outdated", "tender", "ignorant",
		"practical", "scared", "slender", "graceful", "shy", "earnest", "delicate", "fragile",
		"green", "pink", "blue", "terrible", "beautiful", "secret", "mystery", "vapor", "glinting", "shimmering",
		"glimmering", "dull", "dark", "empty", "hollow", "blank", "deafening", "quiet", "stale", "damp", "wet",
		"dry", "thoughtful", "fractured", "violent", "calm", "scared", "frightened",
	},

	{"machines", "students", "desks", "hammers", "loves", "friendships", "teachers", "pianists", "roads", "bedrooms", "humans",
		"computers", "towns", "cities", "cultures", "circles", "rocks", "papers", "shards", "crystals", "keyboards",
		"waves",
	},

	{"orbit", "cry", "hide", "darken", "tan", "profit", "hum", "vibrate", "wander", "chatter", "pose", "shatter", "dissolve", "rotate",
		"hamper", "fall", "bend", "suffer", "remember", "hope", "dream", "blast", "delete", "describe",
		"report", "confuse", "destroy", "slide", "create", "cycle", "grow", "eclipse", "elide", "vocalize",
		"permit", "emit", "evolve", "enjoy",
	},

	{"under", "then", "after", "before", "as if", "like", "by", "above", "narrows", "thins", "shrinks", "blossoms", "through", "amongst", "repeatedly", "one", "my", "our", "their", "his",
		"her", "your", "you", "someone", "anyone", "cleanly through loops of", "tiny", "emanating from the",
	},

	{"the", "antiquated", "mathematical", "mystical", "sad", "sorry", "clever", "some", "thoughtful", "dreadful", "one", "unimaginable", "endless", "finite", "restless", "sleepy", "evening", "sad", "vacant",
		"morning", "loves", "misses", "wonders", "appreciates", "fears", "quietly",
	},

	{"face", "covers", "blankets", "expression", "tone", "impression", "science", "suspicions", "algorithms", "words", "papers", "thoughts", "touches", "breaths", "ideas", "notions", "dreams", "fabric", "clothes",
		"sky", "language", "grammar", "memories", "nights", "earth", "planets", "moons", "sky", "desert",
		"ferns", "faucets", "tables", "memories", "teachers", "soldiers", "workers", "children", "landscape", "arm chair", "legends", "tales", "trash pile", "stories", "veils", "valleys", "peaks",
		"mountain tops", "fruits", "trees", "vines", "birds", "storylines", "embers", "civilization", "culture",
	},
}

var wordsByte [][][]byte

func initWordsByte() {
	wordsByte = make([][][]byte, len(words))
	for i := 0; i < len(words); i++ {
		curr := words[i]
		wordsByte[i] = make([][]byte, len(curr))
		for j := 0; j < len(curr); j++ {
			word := curr[j]
			//wordsByte[i][j]=make([]byte, len(word))
			wordsByte[i][j] = []byte(word)
		}
	}
}
