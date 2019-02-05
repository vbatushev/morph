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
	"bufio"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var extendedTestCases = []struct {
	word string
	want [3][]string
}{
	{"смотри-ка", [3][]string{
		{"смотри-ка"},
		{"смотреть-ка"},
		{"VERB,impf,tran sing,impr,excl"},
	}},
	{"по-западному", [3][]string{
		{"по-западному"},
		{"по-западному"},
		{"ADVB"},
	}},
	{"по-черному", [3][]string{
		{"по-чёрному"},
		{"по-чёрному"},
		{"ADVB"},
	}},
	{"псевдокошка", [3][]string{
		{"псевдокошка", "псевдокошка"},
		{"псевдокошка", "псевдокошка"},
		{"NOUN,anim,femn sing,nomn", "NOUN,inan,femn sing,nomn"},
	}},
	{"интернет-магазин", [3][]string{
		{"интернет-магазин", "интернет-магазин"},
		{"интернет-магазин", "интернет-магазин"},
		{"NOUN,inan,masc sing,nomn", "NOUN,inan,masc sing,accs"},
	}},
	{"человек-гора", [3][]string{
		{"человек-гора", "человек-гора", "человек-гора", "человек-гора"},
		{"человек-гора", "человек-гора", "человек-гор", "человек-гор"},
		{"NOUN,anim,masc sing,nomn", "NOUN,inan,femn sing,nomn", "NOUN,anim,masc,Name sing,gent", "NOUN,anim,masc,Name sing,accs"},
	}},
	{"байткод", [3][]string{
		{"байткод", "байткод", "байткод"},
		{"байткод", "байткод", "байткода"},
		{"NOUN,inan,masc sing,nomn", "NOUN,inan,masc sing,accs", "NOUN,inan,femn plur,gent"},
	}},
	{"бутявкать", [3][]string{
		{"бутявкать", "бутявкать", "бутявкать", "бутявкать", "бутявкать"},
		{"бутявкать", "бутявкатя", "бутявкатя", "бутявкатя", "бутявкать"},
		{"INFN,impf,intr", "NOUN,anim,femn,Name sing,voct,Infr", "NOUN,anim,femn,Name plur,gent", "NOUN,anim,femn,Name plur,accs", "INFN,perf,intr"},
	}},
}

var pymorphy2NormsTests = []struct {
	word  string
	norms []string
}{
	{"кошка", []string{"кошка"}},
	{"кошке", []string{"кошка"}},

	{"стали", []string{"стать", "сталь"}},

	{"наистарейший", []string{"старый"}},

	{"котёнок", []string{"котёнок"}},
	{"котенок", []string{"котёнок"}},
	{"тяжелый", []string{"тяжёлый"}},
	{"легок", []string{"лёгкий"}},

	{"она", []string{"она"}},
	{"ей", []string{"она"}},
	{"я", []string{"я"}},
	{"мне", []string{"я"}},

	{"наиневероятнейший", []string{"вероятный"}},
	{"лучший", []string{"хороший"}},
	{"наилучший", []string{"хороший"}},
	{"человек", []string{"человек"}},
	{"люди", []string{"человек"}},

	{"клюеву", []string{"клюев"}},
	{"клюева", []string{"клюев"}},
	{"иванович", []string{"иванович"}},
	{"ивановичу", []string{"иванович"}},

	{"гулял", []string{"гулять"}},
	{"гуляла", []string{"гулять"}},
	{"гуляет", []string{"гулять"}},
	{"гуляют", []string{"гулять"}},
	{"гуляли", []string{"гулять"}},
	{"гулять", []string{"гулять"}},

	{"гуляющий", []string{"гулять"}},
	{"гулявши", []string{"гулять"}},
	{"гуляя", []string{"гулять"}},
	{"гуляющая", []string{"гулять"}},
	{"загулявший", []string{"загулять"}},

	{"красивый", []string{"красивый"}},
	{"красивая", []string{"красивый"}},
	{"красивому", []string{"красивый"}},
	{"красивые", []string{"красивый"}},

	{"действие", []string{"действие"}},

	{"псевдокошка", []string{"псевдокошка"}},
	{"псевдокошкой", []string{"псевдокошка"}},

	{"сверхнаистарейший", []string{"сверхстарый"}},
	{"сверхнаистарейший", []string{"сверхстарый"}},
	{"квазипсевдонаистарейшего", []string{"квазипсевдостарый"}},
	{"небесконечен", []string{"небесконечный"}},

	{"мегакоту", []string{"мегакот"}},
	{"мегасверхнаистарейшему", []string{"мегасверхстарый"}},

	{"триждычерезпилюлюокнами", []string{"триждычерезпилюлюокно"}},
	{"разквакались", []string{"разквакаться"}},
	{"кашиварнее", []string{"кашиварный"}},
	// XXX: pymorphy2 and morph return these parses in different order
	// {"покашиварней", []string{"кашиварный", "покашиварный", "покашиварня"}},
	// {"подкашиварней", []string{"дкашиварный", "подкашиварный", "подкашиварня"}},
	{"депыртаментов", []string{"депыртамент", "депыртаментовый"}},
	{"измохратился", []string{"измохратиться"}},

	{"бутявкой", []string{"бутявка"}},
	{"сапают", []string{"сапать"}},

	// XXX: pymorphy2 and morph return these parses in different order
	// {"кюди", []string{"кюдить", "кюдь", "кюди"}},
}

var pymorphy2ParsesTests = `
# ========= nouns
кошка       кошка       NOUN,inan,femn sing,nomn

# ========= adjectives
хорошему            хороший     ADJF,Qual masc,sing,datv
лучший              хороший     ADJF,Supr,Qual masc,sing,nomn
наиневероятнейший   вероятный   ADJF,Supr,Qual masc,sing,nomn
наистарейший        старый      ADJF,Supr,Qual masc,sing,nomn

# ========= е/ё
котенок     котёнок     NOUN,anim,masc sing,nomn
котёнок     котёнок     NOUN,anim,masc sing,nomn
озера       озеро       NOUN,inan,neut sing,gent
озера       озеро       NOUN,inan,neut plur,nomn

# ========= particle after a hyphen
ей-то               она-то              NPRO,femn,3per,Anph sing,datv
скажи-ка            сказать-ка          VERB,perf,tran sing,impr,excl
измохратился-таки   измохратиться-таки  VERB,perf,intr masc,sing,past,indc

# ========= compound words with hyphen and immutable left
интернет-магазина       интернет-магазин    NOUN,inan,masc sing,gent
pdf-документов          pdf-документ        NOUN,inan,masc plur,gent
аммиачно-селитрового    аммиачно-селитровый ADJF,Qual masc,sing,gent
быстро-быстро           быстро-быстро       ADVB

# ========= compound words with hyphen and mutable left
команд-участниц     команда-участница   NOUN,inan,femn plur,gent
бегает-прыгает      бегать-прыгать      VERB,impf,intr sing,3per,pres,indc
дул-надувался       дуть-надуваться     VERB,impf,tran masc,sing,past,indc

# ПО- (there were bugs for such words in pymorphy 0.5.6)
почтово-банковский  почтово-банковский  ADJF masc,sing,nomn
по-прежнему         по-прежнему         ADVB

# other old bugs
поездов-экспрессов          поезд-экспресс          NOUN,inan,masc plur,gent
подростками-практикантами   подросток-практикант    NOUN,anim,masc plur,ablt
подводников-североморцев    подводник-североморец   NOUN,anim,masc plur,gent

# issue with normal form caching
залом   зал     NOUN,inan,masc sing,ablt

# cities
санкт-петербурга    санкт-петербург     NOUN,inan,masc,Geox sing,gent
ростове-на-дону     ростов-на-дону      NOUN,inan,masc,Sgtm,Geox sing,loct

# ========= non-dictionary adverbs
по-западному        по-западному        ADVB
по-театральному     по-театральному     ADVB
по-воробьиному      по-воробьиному      ADVB

# ============== common lowercased abbreviations

руб     руб     NOUN,inan,masc,Fixd,Abbr plur,gent
млн     млн     NOUN,inan,masc,Fixd,Abbr plur,gent
тыс     тыс     NOUN,inan,femn,Fixd,Abbr plur,gent
ст      ст      NOUN,inan,femn,Fixd,Abbr sing,accs
`

func TestXParse(t *testing.T) {
	for _, tc := range extendedTestCases {
		words, norms, tags := XParse(tc.word)
		if !reflect.DeepEqual(words, tc.want[0]) {
			t.Errorf("XParse(%q): want words %v, got %v", tc.word, tc.want[0], words)
		}
		if !reflect.DeepEqual(norms, tc.want[1]) {
			t.Errorf("XParse(%q): want norms %v, got %v", tc.word, tc.want[1], norms)
		}
		if !reflect.DeepEqual(tags, tc.want[2]) {
			t.Errorf("XParse(%q): want tags %v, got %v", tc.word, tc.want[2], tags)
		}
	}
}

func uniq(ss []string) []string {
	res := ss[:0]
outer:
	for _, s := range ss {
		for _, t := range res {
			if s == t {
				continue outer
			}
		}
		res = append(res, s)
	}
	return res
}

func TestPymorphy2Norms(t *testing.T) {
	for _, tc := range pymorphy2NormsTests {
		_, norms, _ := XParse(tc.word)
		norms = uniq(norms)
		if !reflect.DeepEqual(norms, tc.norms) {
			t.Errorf("XParse(%q): want norms %v, got %v", tc.word, tc.norms, norms)
		}
	}
}

func TestPymorphy2Parses(t *testing.T) {
	s := bufio.NewScanner(strings.NewReader(pymorphy2ParsesTests))
	rSpace := regexp.MustCompile(`\s+`)
	var tests [][]string
	for s.Scan() {
		line := s.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tests = append(tests, rSpace.Split(line, 3))
	}
	for _, tc := range tests {
		words, norms, tags := XParse(tc[0])
		ok := false
		for i, tag := range tags {
			got := []string{words[i], norms[i], tag}
			if reflect.DeepEqual(got[1:], tc[1:]) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("XParse(%q): want parse %v, got [%v, %v, %v", tc[0], tc, words, norms, tags)
		}
	}
}
