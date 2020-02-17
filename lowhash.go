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

	initWordsFixed(prefix, postfix)

	// prime the program with a cutoff
	cutoff := [32]byte{0xff}

	const batch = 2000000 // about every second, modulo how fast your computer is

	x := append(append([][]string{{prefix}}, words...), []string{postfix})
	methodB(x, cutoff, batch)
	//methodA(cutoff,batch)
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

func methodB(words [][]string, cutoff [32]byte, batch int) {
	start := time.Now()
	grammar := words
	custom, errRead := readLines("lowhash.txt")
	if errRead == nil {
		grammar = custom
	}

	printCombos(grammar)
	combinedWords := combineWords(grammar)

	preserveEnd := 3
	if len(combinedWords) >= preserveEnd {
		front := combinedWords[:len(combinedWords)-preserveEnd]
		back := combinedWords[len(combinedWords)-preserveEnd:]
		frontWide := widenTree(front, 1024*128)
		backWide := widenTree(back, 0)
		recombinedTree := append(frontWide, backWide...)
		combinedWords = recombinedTree
	}

	printTreeBreadth(combinedWords)
	wordsUnfixed := initWordsUnfixed(combinedWords)
	i := 0
	b := 0
	selWords := genRands(wordsUnfixed)

	carries := len(selWords)
	fmt.Printf("random sentence starting point:\n%q\n\n", selectedWordsToStr(selWords, 0, wordsUnfixed))
	var emptyDigest sha256.Sha256Digest
	emptyDigest.Reset()
	var hash [32]byte
	digests := make([][]sha256.Sha256Digest, len(wordsUnfixed)-1)
	var wcnt = len(wordsUnfixed)
	for {
		newWords := false
		for w := wcnt - 1 - carries; w < wcnt; w++ {
			currWords := wordsUnfixed[w]
			if w < wcnt-1 {
				digests[w] = make([]sha256.Sha256Digest, len(currWords))
			}
			for d := 0; d < len(currWords); d++ {
				var currentDigest *sha256.Sha256Digest
				if w == 0 {
					currentDigest = &emptyDigest
				} else {
					currentDigest = &digests[w-1][selWords[w-1]]
				}
				if w < wcnt-1 {
					digests[w][d] = *currentDigest
					digests[w][d].Write(wordsUnfixed[w][d])
				} else {
					cd := *currentDigest
					cd.Write(wordsUnfixed[w][d])
					hash = cd.CheckSum()

					if leftLess(hash, cutoff) {
						s := selectedWordsToStr(selWords, d, wordsUnfixed)
						fmt.Printf("potential found %s\n", s)
						cutoff = postB(s, cutoff)
						newWords = true
					}
					i = i + 1

				}
			}

		}
		if newWords {
			selWords = genRands(wordsUnfixed)
			carries = len(selWords)
			fmt.Printf("Jump to new random sentence:\n%q\n\n", selectedWordsToStr(selWords, 0, wordsUnfixed))
		} else {
			carries = addOneB(selWords, wordsUnfixed)
		}

		if i > batch {
			hashesPerSecond := float64(i) / (float64(time.Since(start).Nanoseconds()) / float64(1000000000))
			start = time.Now()
			fmt.Printf("%3dk hashes per second\n", int64(hashesPerSecond/1000))

			// avoid an int overflow
			i = 0

			if b%10 == 0 {
				fmt.Printf("current position: %q\n", selectedWordsToStr(selWords, 0, wordsUnfixed))
			}
			b = b + 1
		}

	}

}

func methodA(cutoff [32]byte, batch int) {
	start := time.Now()
	var err error
	var arr [256]byte
	var rands [8]int
	n := 0
	rands = genEightRands()
	arr, n = randSentenceFixedArrFasterRands(rands)
	var xx sha256.Sha256Digest
	digest := &xx
	digest.Reset()
	var hash [32]byte
	for i := 0; true; i++ {

		if rands[7] == 0 {
			arr, n = randSentenceFixedArrFasterRands(rands)
			digest.Reset()
			digest.Write(arr[:n])

		}

		// make a copy of the digest
		d0 := *digest
		d0.Write(wordsFixed[7][rands[7]])
		hash = d0.CheckSum()

		if leftLess(hash, cutoff) {

			cutoff, err = post(string(arr[:n])+string(wordsFixed[7][rands[7]]), cutoff)
			if err != nil {
				fmt.Println("Quitting, err " + err.Error())
				return
			}
		}
		rands = addOne(rands)
		if i > batch {

			hashesPerSecond := float64(batch) / (float64(time.Since(start).Nanoseconds()) / float64(1000000000))
			start = time.Now()
			fmt.Printf("%dk hashes per second \n", int64(hashesPerSecond/1000))

			// avoid an int overflow
			i = 0
			// get fresh rands
			rands = genEightRands()
			rands[7] = 0
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

func selectedWordsToStr(selectedWords []int, lastWord int, wordsUnfixed [][][]byte) string {
	//xs := append(selectedWords, lastWord)
	s := ""
	for i := 0; i < len(selectedWords); i++ {
		s += string(wordsUnfixed[i][selectedWords[i]])
	}
	s += string(wordsUnfixed[len(selectedWords)][lastWord])
	return s
}

func postB(s string, cutoff [32]byte) (newCuttoff [32]byte) {

	newCuttoff, err := post(s, cutoff)
	if err != nil {
		fmt.Print("error while posting: %s\n", err.Error())
	}
	return
}

func post(s string, atleast [32]byte) (atleastOut [32]byte, fail error) {

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

func stringToHexBytes(s string) (dout [32]byte) {
	data, _ := hex.DecodeString(s)

	copy(dout[:], data)
	return dout
}

var spaceByte = []byte(" ")[0]

const sixtyfourToEigth int64 = 281474976710656

func genEightRands() (r [8]int) {
	var x = rand.Int63n(sixtyfourToEigth)
	r[0] = int(x & 0x3F)
	x = x >> 5
	r[1] = int(x & 0x3F)
	x = x >> 5
	r[2] = int(x & 0x3F)
	x = x >> 5
	r[3] = int(x & 0x3F)
	x = x >> 5
	r[4] = int(x & 0x3F)
	x = x >> 5
	r[5] = int(x & 0x3F)
	x = x >> 5
	r[6] = int(x & 0x3F)
	x = x >> 5
	r[7] = int(x & 0x3F)
	return
}

func genRands(wordsUnfixed [][][]byte) (r []int) {
	r = make([]int, len(wordsUnfixed)-1)
	for w := 0; w < len(wordsUnfixed)-1; w++ {
		r[w] = rand.Intn(len(wordsUnfixed[w]))
	}
	return
}

func addOne(randIn [8]int) [8]int {

	for pos := 7; pos >= 0; pos = pos - 1 {
		randIn[pos] = randIn[pos] + 1
		if randIn[pos] >= len(words[pos]) {
			randIn[pos] = 0
		} else {
			return randIn
		}
	}
	return randIn
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

func randSentenceFixedArrFasterRands(eightRands [8]int) (arr [256]byte, n int) {
	n = 0

	for i := 0; i < len(wordsFixed)-1; i++ {
		r := eightRands[i]

		for k := 0; k < len(wordsFixed[i][r]); k++ {
			arr[n+k] = wordsFixed[i][r][k]
		}

		n += len(wordsFixed[i][r])
	}

	return arr, n
}

//var sha =  nativeSha.New() //wmSha.New256()

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

var words = [][]string{{
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

var wordsFixed [8][64][]byte

func initWordsFixed(prefix, postfix string) {
	i := 0
	curr := words[i]

	for j := 0; j < 64; j++ {
		word := curr[j%len(curr)]

		wordsFixed[i][j] = []byte(prefix + word)
	}

	for i = 1; i < 7; i++ {
		curr = words[i]

		for j := 0; j < 64; j++ {
			word := curr[j%len(curr)]

			wordsFixed[i][j] = []byte(word)
		}
	}

	i = 7
	curr = words[i]

	for j := 0; j < 64; j++ {
		word := curr[j%len(curr)]

		wordsFixed[i][j] = []byte(word + postfix)
	}

}

func initWordsUnfixed(wordPositions [][]string) (wordPositionsBytes [][][]byte) {
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
