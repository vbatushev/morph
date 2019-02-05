// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package morph

import (
	"encoding/binary"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var particlesAfterHyphen = []string{
	"-то",
	"-ка",
	"-таки",
	"-де",
	"-тко",
	"-тка",
	"-с",
	"-ста",
}

var knownPrefixes = []string{
	"авиа",
	"авто",
	"аква",
	"анти-",
	"анти",
	"антропо",
	"арт-",
	"арт",
	"архи",
	"астро",
	"аудио",
	"аэро",
	"без",
	"бес",
	"био",
	"вело",
	"взаимо",
	"видео",
	"вице-",
	"вне",
	"внутри",
	"вперед",
	"впереди",
	"гекто",
	"гелио",
	"гео",
	"гетеро",
	"гига",
	"гигро",
	"гипер",
	"гипо",
	"гомо",
	"дву",
	"двух",
	"де",
	"дез",
	"дека",
	"деци",
	"дис",
	"до",
	"евро",
	"за",
	"зоо",
	"интер",
	"инфра",
	"квази-",
	"квази",
	"кило",
	"кино",
	"контр-",
	"контр",
	"космо-",
	"космо",
	"крипто",
	"лейб-",
	"лже-",
	"лже",
	"макро",
	"макси-",
	"макси",
	"мало",
	"мега",
	"медиа-",
	"медиа",
	"меж",
	"мета-",
	"мета",
	"метео",
	"метро",
	"микро",
	"милли",
	"мини-",
	"мини",
	"много",
	"моно",
	"мото",
	"мульти",
	"нано",
	"нарко",
	"не",
	"небез",
	"недо",
	"нейро",
	"нео",
	"низко",
	"обер-",
	"обще",
	"одно",
	"около",
	"орто",
	"палео",
	"пан",
	"пара",
	"пента",
	"пере",
	"пиро",
	"поли",
	"полу",
	"порно",
	"после",
	"пост-",
	"пост",
	"пра-",
	"пра",
	"пред",
	"пресс-",
	"противо-",
	"противо",
	"прото",
	"псевдо-",
	"псевдо",
	"радио",
	"разно",
	"ре",
	"ретро-",
	"ретро",
	"само",
	"санти",
	"сверх-",
	"сверх",
	"спец",
	"суб",
	"супер-",
	"супер",
	"супра",
	"теле",
	"тетра",
	"топ-",
	"транс-",
	"транс",
	"ультра",
	"унтер-",
	"штаб-",
	"экзо",
	"эко",
	"эконом-",
	"экс-",
	"экс",
	"экстра-",
	"экстра",
	"электро",
	"эндо",
	"энерго",
	"этно",
}

var nonproductiveGrammemes = []string{
	"NUMR",
	"NPRO",
	"PRED",
	"PREP",
	"CONJ",
	"PRCL",
	"INTJ",
	"Apro",
}

func init() {
	sort.Slice(knownPrefixes, func(i, j int) bool {
		d := len(knownPrefixes[i]) - len(knownPrefixes[j])
		if d != 0 {
			return d > 0
		}
		return knownPrefixes[i] < knownPrefixes[j]
	})
}

func productive(tag string) bool {
	for _, g := range nonproductiveGrammemes {
		if strings.Contains(tag, g) {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func wordSplits(s string, minRemainder, maxPrefixLen int) [][]string {
	rr := []rune(s)
	n := min(maxPrefixLen, len(rr)-minRemainder)
	if n <= 0 {
		return nil
	}
	ss := make([][]string, 0, n)
	for i := 1; i <= n; i++ {
		ss = append(ss, []string{
			string(rr[:i]),
			string(rr[i:]),
		})
	}
	return ss
}

func split5(s string) [][2]string {
	var splits [][2]string
	n := len(s)
	for i := 0; i < 5; i++ {
		_, size := utf8.DecodeLastRuneInString(s[:n])
		n -= size
		if n == 0 {
			break
		}
		splits = append(splits, [2]string{s[:n], s[n:]})
	}
	return splits
}

func emptyUnlessFeature(s string) string {
	switch s {
	case "loc1":
		return "loct"
	case "gen1":
		return "gent"

		// parts of speech
	case "NOUN", "ADJF", "ADJS", "COMP", "VERB", "INFN", "PRTF", "PRTS", "GRND", "NUMR", "ADVB", "NPRO", "PRED", "PREP", "CONJ", "PRCL", "INTJ":
		fallthrough

		// numbers
	case "sing", "plur":
		fallthrough

		// cases
	case "nomn", "gent", "datv", "accs", "ablt", "loct", "voct", "gen2", "acc2", "loc2":
		fallthrough

		// persons
	case "1per", "2per", "3per":
		fallthrough

		// tenses
	case "pres", "past", "futr":
		return s

	}
	return ""
}

var rFeature = regexp.MustCompile(`[^ ,]+`)

func similarityFeatures(tag string) string {
	return rFeature.ReplaceAllStringFunc(tag, emptyUnlessFeature)
}

// XParse analyzes the word (which might not be in the dictionary)
// and returns three slices of the same length.
// Each triple (words[i], norms[i], tags[i]) represents an analysis, where:
// - words[i] is the word with the letter ё fixed;
// - norms[i] is the normal form of the word;
// - tags[i] is the grammatical tag, consisting of the word's grammemes.
// If the word is in the dictionary, XParse is equivalent to Parse.
// Otherwise it tries several other analyzers to analyze the unknown word.
func XParse(word string) (words, norms, tags []string) {
	word = strings.ToLower(word)
	words, norms, tags = Parse(word)
	if len(words) > 0 {
		return words, norms, tags
	}

	containsHyphen := strings.IndexByte(word, '-') != -1

	// try to strip a particle after the hyphen, e.g. смотри-ка -> смотри + ка
	// (HyphenSeparatedParticleAnalyzer in pymorphy2)
	if containsHyphen {
		for _, suffix := range particlesAfterHyphen {
			if !strings.HasSuffix(word, suffix) {
				continue
			}
			unsuffixed := strings.TrimSuffix(word, suffix)
			words, norms, tags := XParse(unsuffixed)
			if len(words) > 0 {
				for i := range words {
					words[i] += suffix
					norms[i] += suffix
				}
				return words, norms, tags
			}
		}
	}

	nRunes := utf8.RuneCountInString(word)

	// parse adverbs starting with по-, e.g. по-западному
	// (HyphenAdverbAnalyzer in pymorphy2)
	if nRunes >= 5 && strings.HasPrefix(word, "по-") {
		words, _, tags := XParse(word[5:])
		for i, tag := range tags {
			if !strings.HasPrefix(tag, "ADJF") ||
				!strings.Contains(tag, "sing,datv") {
				continue
			}
			w := "по-" + words[i]
			return []string{w}, []string{w}, []string{"ADVB"}
		}
	}

	// parse words starting with known prefixes, e.g. псевдокошка -> (псевдо) + кошка
	// (KnownPrefixAnalyzer in pymorphy2)
	for _, prefix := range knownPrefixes {
		if !strings.HasPrefix(word, prefix) {
			continue
		}
		unprefixed := strings.TrimPrefix(word, prefix)
		if utf8.RuneCountInString(unprefixed) < 3 {
			continue
		}
		ws, ns, ts := XParse(unprefixed)
		for i, tag := range ts {
			if !productive(tag) {
				continue
			}
			words = append(words, prefix+ws[i])
			norms = append(norms, prefix+ns[i])
			tags = append(tags, ts[i])
		}
	}
	if len(words) > 0 {
		return words, norms, tags
	}

	// parse word by parsing its hyphen-separated parts, e.g.
	// интернет-магазин -> "интернет-" + магазин
	// человек-гора -> человек + гора
	// (HyphenatedWordsAnalyzer in pymorphy2)
	if containsHyphen && strings.Count(word, "-") == 1 &&
		!strings.HasPrefix(word, "-") && !strings.HasSuffix(word, "-") {

		parts := strings.SplitN(word, "-", 2)
		left, right := parts[0], parts[1]
		lwords, lnorms, ltags := XParse(left)
		rwords, rnorms, rtags := XParse(right)
		rightFeatures := make([]string, len(rtags))
		for i, tag := range rtags {
			rightFeatures[i] = similarityFeatures(tag)
		}
		for i, tag := range ltags {
			leftFeat := similarityFeatures(tag)
			for j := range rtags {
				if leftFeat != rightFeatures[j] {
					continue
				}
				words = append(words, lwords[i]+"-"+rwords[j])
				norms = append(norms, lnorms[i]+"-"+rnorms[j])
				tags = append(tags, tag)
			}
		}
		for i, tag := range rtags {
			words = append(words, left+"-"+rwords[i])
			norms = append(norms, left+"-"+rnorms[i])
			tags = append(tags, tag)
		}
		if len(words) > 0 {
			return words, norms, tags
		}
	}

	// try parsing only the suffix (with restrictions on prefix and suffix lengths), e.g.
	// байткод -> (байт) + код
	// (UnknownPrefixAnalyzer in pymorphy2)
	for _, split := range wordSplits(word, 3, 5) {
		prefix, unprefixed := split[0], split[1]
		ws, ns, ts := Parse(unprefixed)
		for i, tag := range ts {
			if !productive(tag) {
				continue
			}
			words = append(words, prefix+ws[i])
			norms = append(norms, prefix+ns[i])
			tags = append(tags, ts[i])
		}
	}

	// parse the word by checking how the words with similar suffixes are parsed, e.g.
	// бутявкать -> ...вкать
	// (KnownSuffixAnalyzer in pymorphy2)
	if nRunes >= 4 {
		splits := split5(word)
		for id, prefix := range prefixes {
			if !strings.HasPrefix(word, prefix) {
				continue
			}
			totalCount := 0
			dawg := predictionDAWGs[id]
			for i := len(splits) - 1; i >= 0; i-- {
				sp := splits[i]
				wordStart, wordEnd := sp[0], sp[1]
			sloop:
				for _, it := range dawg.similarItems(wordEnd) {
					for _, v := range it.values {
						count := int(binary.BigEndian.Uint16(v))
						paraNum := int(binary.BigEndian.Uint16(v[2:]))
						para := paradigms[paraNum]
						index := int(binary.BigEndian.Uint16(v[4:]))

						prefix, suffix, tag := prefixSuffixTag(para, index)
						if !productive(tag) {
							continue
						}

						totalCount += count

						word := wordStart + it.key
						norm := word
						if index != 0 {
							stem := strings.TrimPrefix(norm, prefix)
							stem = strings.TrimSuffix(stem, suffix)
							pr, su, _ := prefixSuffixTag(para, 0)
							norm = pr + stem + su
						}

						for i, t := range tags {
							if t == tag && words[i] == word && norms[i] == norm {
								continue sloop
							}
						}

						words = append(words, word)
						norms = append(norms, norm)
						tags = append(tags, tag)
					}
				}
				if totalCount > 1 {
					break
				}
			}
		}
	}

	return words, norms, tags
}
