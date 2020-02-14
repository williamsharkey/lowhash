package main

import (
	nativeSha "crypto/sha256"
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

	fmt.Println()
	fmt.Println(" +-----------+")
	fmt.Println(" |  lowhash  |")
	fmt.Println(" +-----------+")
	fmt.Println()
	fmt.Println(" https://lowhash.com")
	fmt.Println()
	fmt.Printf("prefix: %q\n", prefix)
	fmt.Printf("postfix: %q\n", postfix)
	fmt.Println("to set a prefix/postfix, run with arguments, eg: lowhash \"my prefix\" \"my postfix\"")
	fmt.Println()
	start := time.Now()

	// prime the program with a cutoff
	atleast, err := post("a new paper suffers as if the thoughts", "")

	const batch = 4000000

	for i := 1; true; i++ {
		s := randSentence(prefix, postfix)

		hash := hash(s)

		if hash < atleast {
			atleast, err = post(s, atleast)
			if err != nil {
				fmt.Println("Quitting, err " + err.Error())
				return
			}
		}

		if i%batch == 0 {
			hashesPerSecond := float64(batch) / (float64(time.Since(start).Nanoseconds()) / float64(1000000000))
			start = time.Now()
			fmt.Printf("%dk hashes per second \n", int64(hashesPerSecond/1000))
		}

	}

}

var cutoffRegex, _ = regexp.Compile(".*cutoff: (.*)")
var msgRegex, _ = regexp.Compile(".*msg: (.*)")
var rankRegex, _ = regexp.Compile(".*rank: (.*)")

func post(s string, atleast string) (atleastOut string, fail error) {
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
				atleastOut = cutoffRegex.FindStringSubmatch(bodyString)[1]
				fmt.Println("Set cutoff to " + atleastOut[0:8] + " " + atleastOut[8:16])
			} else if strings.HasPrefix(bodyString, "\nrank: ") {
				atleastOut = cutoffRegex.FindStringSubmatch(bodyString)[1]
				rank := rankRegex.FindStringSubmatch(bodyString)[1]
				msg := msgRegex.FindStringSubmatch(bodyString)[1]
				fmt.Println("Excellent! You placed " + rank + " with: \"" + msg + "\"")
				fmt.Println(" Update cutoff to " + atleastOut[0:8] + " " + atleastOut[8:16])
			} else if strings.HasPrefix(bodyString, "\nalready present") {
				atleastOut = cutoffRegex.FindStringSubmatch(bodyString)[1]
				fmt.Println("Set cutoff to " + atleastOut[0:8] + " " + atleastOut[8:16])
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

func hash(x string) string {
	sha.Reset()
	io.WriteString(sha, x)
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func randSentence(pre, post string) string {
	sentence := pre
	for i := 0; i < len(words); i++ {
		var j = len(words[i])
		var r = rand.Intn(j)
		if i == 0 {
			sentence += strings.ToUpper(words[i][r][0:1]) + words[i][r][1:]
		} else {
			sentence += " " + words[i][r]
		}

	}
	return sentence + post
}

var sha = nativeSha.New()

func init() {
	rand.Seed(time.Now().Unix())

}

var words = [][]string{{
	"capture", "dispense", "eject", "imagine", "be", "do", "have", "become", "once", "if", "imagine if", "suppose that", "I know", "don't fear",
	"because", "therefore", "as though", "we think", "I believe", "you hope"},

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
