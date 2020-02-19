package main

import (
	"bufio"
	//nativeSha "crypto/sha256"
	"github.com/williamsharkey/sha256-simd"
	//"github.com/minio/sha256-simd"
	//wmSha "github.com/williamsharkey/sha256"
	"encoding/hex"
	"fmt"
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
	postfix := ""

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

	const batch = 2000000 * 5 // about every 5 seconds, modulo how fast your computer is

	grammar := generateWords(defaultWords, prefix, postfix)

	digestTree(grammar, batch)
}

func printCombos(words [][]string) {
	combos := 1
	for i := 0; i < len(words); i++ {
		combos = combos * len(words[i])
	}
	fmt.Printf("There are %d possible sententces\n", combos)
}

func printTreeBreadth(words [][]string) {
	s := ""
	for i := 0; i < len(words)-1; i++ {
		s += fmt.Sprintf("%d, ", len(words[i]))
	}
	s += fmt.Sprintf("%d", len(words[len(words)-1]))

	fmt.Printf("Tree Breadth: [%s]\n", s)
}

func generateWords(defaultGrammar [][]string, prefix, postfix string) Grammar {

	grammar := defaultGrammar
	custom, errRead := readLines("grammar.txt")
	if errRead == nil {
		grammar = custom
	}
	grammar = append(append([][]string{{prefix}}, grammar...), []string{postfix})

	printCombos(grammar)
	combinedWords := combineWords(grammar)

	preserveEnd := 1
	if len(combinedWords) >= preserveEnd {
		front := combinedWords[:len(combinedWords)-preserveEnd]
		back := combinedWords[len(combinedWords)-preserveEnd:]
		frontWide := widenTree(front, 0) //1024*128)
		backWide := widenTree(back, 0)
		recombinedTree := append(frontWide, backWide...)

		printTreeBreadth(recombinedTree)
		return strsToByte(recombinedTree)
	} else {

		printTreeBreadth(combinedWords)
		return strsToByte(combinedWords)

	}

}

type Grammar [][][]byte

func (g Grammar) Str(selWords []int, lastWord int) string {
	a := append(selWords, lastWord)
	s := ""
	for i := 0; i < len(a); i++ {
		s += string(g[i][a[i]])
	}
	return s
}

func digestTree(grammar Grammar, batch int) {
	start := time.Now()
	i := 0
	b := 0
	selWords := genRands(grammar)
	carries := len(selWords)
	initStr := grammar.Str(selWords, 0)
	cutoff := post(initStr, sha256.Sum256([]byte(initStr)))

	fmt.Printf("Start search with:\n%q\n\n", initStr)
	var emptyDigest sha256.Sha256Digest
	emptyDigest.Reset()

	digests := make([]sha256.Sha256Digest, len(grammar)-1)
	var wcnt = len(grammar)
	for {
		newWords := false
		for w := wcnt - 1 - carries; w < wcnt; w++ {

			var currentDigest *sha256.Sha256Digest
			if w == 0 {
				currentDigest = &emptyDigest
			} else {
				currentDigest = &digests[w-1]
			}
			if w < wcnt-1 {
				digests[w] = *currentDigest
				digests[w].Write(grammar[w][selWords[w]])

			} else {

				for j := 0; j < len(grammar[w]); j++ {

					cd := *currentDigest
					cd.Write(grammar[w][j])

					if cd.CheckSumLessThanOrEqual(cutoff) {
						s := grammar.Str(selWords, j)
						fmt.Printf("potential found %s\n", s)
						//print title
						fmt.Printf("\033]0;%q\007", s)
						cutoff = post(s, cutoff)
						newWords = true
					}
					i = i + 1
				}
			}

		}
		if i > batch {
			hashesPerSecond := float64(i) / (float64(time.Since(start).Nanoseconds()) / float64(1000000000))
			start = time.Now()
			fmt.Printf("%3dk hashes per second\n", int64(hashesPerSecond/1000))

			// avoid an int overflow
			i = 0

			if b%10 == 0 {
				fmt.Printf("current position: %q\n", grammar.Str(selWords, 0))
			}
			b = b + 1
			newWords = true
		}

		if newWords {
			selWords = genRands(grammar)
			carries = len(selWords)
			fmt.Printf("Jump to:\n%q\n\n", grammar.Str(selWords, 0))
		} else {
			carries = addOneB(selWords, grammar)
		}

	}

}

func leftLess(a, b [32]byte) bool {
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

func post(s string, cutoff [32]byte) (newCutoff [32]byte) {

	newCutoff = cutoff
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
				fmt.Printf("err reading response: %s\n", err2)

			} else {
				bodyString := string(bodyBytes)
				if strings.HasPrefix(bodyString, "\nhash does not rank") {
					a := cutoffRegex.FindStringSubmatch(bodyString)[1]
					newCutoff = stringToHexBytes(a)
					fmt.Println("Set cutoff to " + a[0:8] + " " + a[8:16])
				} else if strings.HasPrefix(bodyString, "\nrank: ") {
					b := cutoffRegex.FindStringSubmatch(bodyString)[1]
					newCutoff = stringToHexBytes(b)
					rank := rankRegex.FindStringSubmatch(bodyString)[1]
					msg := msgRegex.FindStringSubmatch(bodyString)[1]
					fmt.Println("Excellent! You placed " + rank + " with: \"" + msg + "\"")
					fmt.Println(" Update cutoff to " + b[0:8] + " " + b[8:16])
				} else if strings.HasPrefix(bodyString, "\nalready present") {
					c := cutoffRegex.FindStringSubmatch(bodyString)[1]
					newCutoff = stringToHexBytes(c)
					fmt.Println("Set cutoff to " + c[0:8] + " " + c[8:16])
				} else {
					fmt.Print("unknown response")
					fmt.Print(bodyString)
				}
			}
		} else {
			fmt.Printf("status code %v\n", resp.StatusCode)
			fail := fmt.Errorf("%d", resp.StatusCode)
			if fail != nil {
				fmt.Printf("error while posting: %s\n", err.Error())
			}
		}
	}
	return
}

func stringToHexBytes(s string) (dout [32]byte) {
	data, _ := hex.DecodeString(s)

	copy(dout[:], data)
	return dout
}

func genRands(wordsUnfixed [][][]byte) (r []int) {
	r = make([]int, len(wordsUnfixed)-1)
	for w := 0; w < len(wordsUnfixed)-1; w++ {
		r[w] = rand.Intn(len(wordsUnfixed[w]))
	}
	return
}

func addOneB(selWords []int, wordsUnfixed [][][]byte) (carries int) {

	for pos := len(selWords) - 1; pos >= 0; pos = pos - 1 {
		selWords[pos] = selWords[pos] + 1
		if selWords[pos] >= len(wordsUnfixed[pos]) {
			carries = carries + 1
			selWords[pos] = 0
		} else {
			return
		}
	}
	return
}

func init() {
	rand.Seed(time.Now().Unix())
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())

		for i := 0; i < len(f); i++ {
			f[i] = strings.Replace(f[i], "_", " ", -1)
		}
		lines = append(lines, f)
	}
	return lines, scanner.Err()
}

var defaultWords = [][]string{{
	"Capture", "Dispense", "Eject", "Imagine", "Be", "Do", "Have", "Become", "Once", "If", "Imagine if", "Suppose that",
	"I know", "Don't fear", "Because", "Therefore", "As though", "We think", "I believe", "You hope"},
	{" "},
	{"those", "my", "her", "his", "what", "our", "when", "that", "some", "one", "a", "some", "the", "a", "the"},
	{" "},
	{"lost", "hidden", "sorry", "embodied", "intrinsic", "electronic", "new", "outdated", "tender", "ignorant",
		"practical", "scared", "slender", "graceful", "shy", "earnest", "delicate", "fragile",
		"green", "pink", "blue", "terrible", "beautiful", "secret", "mystery", "vapor", "glinting", "shimmering",
		"glimmering", "dull", "dark", "empty", "hollow", "blank", "deafening", "quiet", "stale", "damp", "wet",
		"dry", "thoughtful", "fractured", "violent", "calm", "scared", "frightened",
	},
	{" "},
	{"machines", "students", "desks", "hammers", "loves", "friendships", "teachers", "pianists", "roads", "bedrooms", "humans",
		"computers", "towns", "cities", "cultures", "circles", "rocks", "papers", "shards", "crystals", "keyboards",
		"waves",
	},
	{" "},
	{"orbit", "cry", "hide", "darken", "tan", "profit", "hum", "vibrate", "wander", "chatter", "pose", "shatter", "dissolve", "rotate",
		"hamper", "fall", "bend", "suffer", "remember", "hope", "dream", "blast", "delete", "describe",
		"report", "confuse", "destroy", "slide", "create", "cycle", "grow", "eclipse", "elide", "vocalize",
		"permit", "emit", "evolve", "enjoy",
	},
	{" "},
	{"under", "then", "after", "before", "as if", "like", "by", "above", "narrows", "thins", "shrinks", "blossoms", "through", "amongst", "repeatedly", "one", "my", "our", "their", "his",
		"her", "your", "you", "someone", "anyone", "cleanly through loops of", "tiny", "emanating from the",
	},
	{" "},
	{"the", "antiquated", "mathematical", "mystical", "sad", "sorry", "clever", "some", "thoughtful", "dreadful", "one", "unimaginable", "endless", "finite", "restless", "sleepy", "evening", "sad", "vacant",
		"morning", "loves", "misses", "wonders", "appreciates", "fears", "quietly",
	},

	{" "},

	{"face", "covers", "blankets", "expression", "tone", "impression", "science", "suspicions", "algorithms", "words", "papers", "thoughts", "touches", "breaths", "ideas", "notions", "dreams", "fabric", "clothes",
		"sky", "language", "grammar", "memories", "nights", "earth", "planets", "moons", "sky", "desert",
		"ferns", "faucets", "tables", "memories", "teachers", "soldiers", "workers", "children", "landscape", "arm chair", "legends", "tales", "trash pile", "stories", "veils", "valleys", "peaks",
		"mountain tops", "fruits", "trees", "vines", "birds", "storylines", "embers", "civilization", "culture",
	},

	{".", "?", "!", ""},
}

func strsToByte(wordPositions [][]string) (wordPositionsBytes [][][]byte) {
	wordPositionsBytes = make([][][]byte, len(wordPositions))
	for i := 0; i < len(wordPositions); i++ {
		curr := wordPositions[i]
		wordPositionsBytes[i] = make([][]byte, len(curr))
		for j := 0; j < len(curr); j++ {
			word := curr[j]
			wordPositionsBytes[i][j] = []byte(word)
		}
	}

	return wordPositionsBytes
}

func combineWords(ws [][]string) [][]string {
	c := [][]string{}
	last := ""
	for i := 0; i < len(ws); i++ {

		if len(ws[i]) == 1 {
			//copyLast=true
			last = last + ws[i][0]
		} else {
			c = append(c, ws[i])
			for j := 0; j < len(ws[i]); j++ {
				c[len(c)-1][j] = last + c[len(c)-1][j]
			}
			last = ""
		}
	}
	// postfix final entries
	if last != "" {
		if len(c) > 0 {
			for j := 0; j < len(c[len(c)-1]); j++ {
				c[len(c)-1][j] = c[len(c)-1][j] + last
			}
		} else {
			c = append(c, []string{last})
		}
	}
	return c
}

func widenTree(ws [][]string, maxWidth int) [][]string {
	if len(ws) < 2 {
		return ws
	}

	min := len(ws[0]) * len(ws[1])
	minI := 0
	for i := 1; i < len(ws)-1; i++ {
		curr := len(ws[i]) * len(ws[i+1])
		if curr < min {
			min = curr
			minI = i
		}
	}

	if min > maxWidth {
		return ws
	}
	outArr := make([][]string, len(ws)-1)
	for i := 0; i < minI; i++ {
		outArr[i] = make([]string, len(ws[i]))
		for j := 0; j < len(ws[i]); j++ {
			outArr[i][j] = ws[i][j]
		}
	}

	outArr[minI] = make([]string, min)
	for i := 0; i < len(ws[minI]); i++ {
		for j := 0; j < len(ws[minI+1]); j++ {
			outArr[minI][j+i*(len(ws[minI+1]))] = ws[minI][i] + ws[minI+1][j]
		}
	}

	for i := minI + 1; i < len(outArr); i++ {
		outArr[i] = make([]string, len(ws[i+1]))
		for j := 0; j < len(ws[i+1]); j++ {
			outArr[i][j] = ws[i+1][j]
		}
	}
	return widenTree(outArr, maxWidth)

}
