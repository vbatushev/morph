morph [![License](http://img.shields.io/:license-gpl3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.html) [![GoDoc](http://godoc.org/gitlab.com/opennota/morph?status.svg)](http://godoc.org/gitlab.com/opennota/morph)
=====

Морфологический анализатор русского языка, использующий словари [pymorphy2](https://github.com/kmike/pymorphy2).

## Установка

Пакет:

    go get -u github.com/vbatushev/morph

Словари:

    pip install --user pymorphy2-dicts-ru

## Использование

``` go
package main
import (
    "fmt"
    "gitlab.com/opennota/morph"
)
func main() {
    // loading the dictionary data
    if err := morph.Init(); err != nil {
        panic(err)
    }
    // parsing
    words, norms, tags := morph.Parse("все")
    for i := range words {
        fmt.Printf("%-4s %-5s %s\n", words[i], norms[i], tags[i])
    }
}
```

Вывод:

    все  весь  ADJF,Subx,Apro plur,nomn
    все  весь  ADJF,Subx,Apro inan,plur,accs
    всё  всё   PRCL
    всё  весь  ADJF,Subx,Apro neut,sing,nomn
    всё  весь  ADJF,Subx,Apro neut,sing,accs
