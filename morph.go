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

// Package morph provides a simple morphological analyzer for Russian language,
// using the compiled dictionaries from pymorphy2.
package morph

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrAlreadyInitialized = errors.New("already initialized")

	prefixes        []string
	suffixes        []string
	tags            []string
	paradigms       [][]uint16
	wordsDAWG       *dawg
	probDAWG        *dawg
	predictionDAWGs []*dawg
)

type parse struct {
	words []string
	norms []string
	tags  []string
	probs []float64
}

func (p *parse) Len() int           { return len(p.words) }
func (p *parse) Less(i, j int) bool { return p.probs[i] > p.probs[j] }
func (p *parse) Swap(i, j int) {
	p.words[i], p.words[j] = p.words[j], p.words[i]
	p.norms[i], p.norms[j] = p.norms[j], p.norms[i]
	p.tags[i], p.tags[j] = p.tags[j], p.tags[i]
	p.probs[i], p.probs[j] = p.probs[j], p.probs[i]
}

// Parse analyzes the (lowercase) word and returns three slices of the same length.
// Each triple (words[i], norms[i], tags[i]) represents an analysis, where:
// - words[i] is the word with the letter ё fixed;
// - norms[i] is the normal form of the word;
// - tags[i] is the grammatical tag, consisting of the word's grammemes.
// The analyzes are sorted by probability (the first one is the most probable).
func Parse(word string) (words, norms, tags []string) {
	if probDAWG == nil {
		panic("not initialized; call Init or InitWith")
	}

	var probs []float64
	hasNonzeroProb := false

	for _, it := range wordsDAWG.similarItems(word) {
		for _, v := range it.values {
			paraNum := int(binary.BigEndian.Uint16(v))
			para := paradigms[paraNum]
			index := int(binary.BigEndian.Uint16(v[2:]))

			prefix, suffix, tag := prefixSuffixTag(para, index)

			norm := it.key
			if index != 0 {
				stem := strings.TrimPrefix(norm, prefix)
				stem = strings.TrimSuffix(stem, suffix)
				pr, su, _ := prefixSuffixTag(para, 0)
				norm = pr + stem + su
			}

			words = append(words, it.key)
			norms = append(norms, norm)
			tags = append(tags, tag)

			prob := float64(probDAWG.Dict.find(word+":"+tag)) / 1e6
			if prob > 0 {
				hasNonzeroProb = true
			}
			probs = append(probs, prob)
		}
	}

	if hasNonzeroProb {
		sort.Stable(&parse{words, norms, tags, probs})
	}

	return words, norms, tags
}

// Init tries to find the path to the installed pymorphy2 dictionaries by invoking python and calls InitWith with the found directory.
func Init() error {
	if probDAWG != nil {
		return ErrAlreadyInitialized
	}

	dir, err := dataPath()
	if err != nil {
		return err
	}
	return InitWith(dir)
}

// InitWith loads the pymorphy2 dictionary data from the given directory.
func InitWith(dir string) error {
	if probDAWG != nil {
		return ErrAlreadyInitialized
	}

	prefixesPath := filepath.Join(dir, "paradigm-prefixes.json")
	suffixesPath := filepath.Join(dir, "suffixes.json")
	tagsPath := filepath.Join(dir, "gramtab-opencorpora-int.json")
	paradigmsPath := filepath.Join(dir, "paradigms.array")
	dawgPath := filepath.Join(dir, "words.dawg")
	probPath := filepath.Join(dir, "p_t_given_w.intdawg")

	var err error

	tags, err = loadStringArray(tagsPath)
	if err != nil {
		return err
	}

	prefixes, err = loadStringArray(prefixesPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		prefixes = []string{"", "по", "наи"}
	}

	suffixes, err = loadStringArray(suffixesPath)
	if err != nil {
		return err
	}

	if err := loadParadigms(paradigmsPath); err != nil {
		return err
	}

	wordsDAWG, err = newDAWG(dawgPath)
	if err != nil {
		return err
	}

	probDAWG, err = newDAWG(probPath)
	if err != nil {
		return err
	}

	predictionDAWGs = nil
	for i := 0; i < len(prefixes); i++ {
		path := filepath.Join(dir, fmt.Sprintf("prediction-suffixes-%d.dawg", i))
		d, err := newDAWG(path)
		if err != nil {
			return err
		}
		predictionDAWGs = append(predictionDAWGs, d)
	}

	return nil
}

func dataPath() (string, error) {
	cmd := exec.Command("python", "-c", "import pymorphy2_dicts_ru as p; print(p.__path__[0])")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	dir := strings.TrimRight(buf.String(), "\r\n")
	return filepath.Join(dir, "data"), nil
}

func loadStringArray(fn string) ([]string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ss []string
	if err := json.NewDecoder(f).Decode(&ss); err != nil {
		return nil, err
	}
	return ss, nil
}

func loadParadigms(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	var paraCount uint16
	if err := binary.Read(f, binary.LittleEndian, &paraCount); err != nil {
		return err
	}

	paradigms = make([][]uint16, 0, paraCount)
	for i := 0; i < int(paraCount); i++ {
		var paraLen uint16
		if err := binary.Read(f, binary.LittleEndian, &paraLen); err != nil {
			return err
		}

		para := make([]uint16, paraLen)
		if err := binary.Read(f, binary.LittleEndian, &para); err != nil {
			return err
		}

		paradigms = append(paradigms, para)
	}

	return nil
}

func prefixSuffixTag(para []uint16, i int) (string, string, string) {
	n := len(para) / 3
	suffixIndex := para[i]
	tagIndex := para[i+n]
	prefixIndex := para[i+2*n]
	return prefixes[prefixIndex], suffixes[suffixIndex], tags[tagIndex]
}
