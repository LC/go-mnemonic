/*Package bip39 is an immutable class that represents a BIP39 Mnemonic code.
  See BIP39 specification for more info: https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
  A Mnemonic code is a a group of easy to remember words used for the generation
  of deterministic wallets. A Mnemonic can be used to generate a seed using
  an optional passphrase, for later generate a HDPrivateKey. */
package bip39

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const bitsInByte = 8
const minEnt = 128
const maxEnt = 256
const multiple = 32
const wordBits = 11

//Mnemonic ...
type Mnemonic struct {
	ent        []byte
	passphrase string
	sentence   string
}

/*NewMnemonicRandom generate a group of easy to remember words
 -- for the generation of deterministic wallets.
use size 128 for a 12 words code.*/
func NewMnemonicRandom(size int, passphrase string) (code *Mnemonic, e error) {
	//we generate ENT count of random bits
	ent, err := generateEntropy(size)
	if err != nil {
		e = err
		return
	}

	code = &Mnemonic{}
	code.ent = ent
	code.passphrase = passphrase

	return
}

//NewMnemonicFromEntropy ...
func NewMnemonicFromEntropy(ent []byte, passphrase string) (code *Mnemonic, e error) {
	code = &Mnemonic{}
	code.ent = ent
	code.passphrase = passphrase
	return
}

//newMnemonicFromSentence ...
func newMnemonicFromSentence(sentence string, passphrase string) (code *Mnemonic, e error) {
	//TODO
	return
}

//GetSentence ...
func (m *Mnemonic) GetSentence() (string, error) {
	if len(m.sentence) != 0 {
		return m.sentence, nil
	}

	// entCS := len(m.ent) * bitsInByte
	// ms := entCS / wordBits

	/*  var bin = '';
	for (var i = 0; i < entropy.length; i++) {
	  bin = bin + ('00000000' + entropy[i].toString(2)).slice(-8);
	}

	bin = bin + Mnemonic._entropyChecksum(entropy);
	if (bin.length % 11 !== 0) {
	  throw new errors.InvalidEntropy(bin);
	}
	var mnemonic = [];
	for (i = 0; i < bin.length / 11; i++) {
	  var wi = parseInt(bin.slice(i * 11, (i + 1) * 11), 2);
	  mnemonic.push(wordlist[wi]);
	} */

	checksum, err := checksumEntropy(m.ent)
	if err != nil {
		return "", err
	}

	ent := append(m.ent, checksum...)

	bin := ""
	for _, b := range ent {
		bin = bin + fmt.Sprintf("%08b", b)
	}

	wordCount := len(bin) / wordBits
	if len(bin)%wordBits != 0 {
		err := fmt.Errorf("internal error, canot divide ENT to %v groups", wordBits)
		return "", err
	}

	groups := make([]int, wordCount)
	var str string
	for i := 0; i < wordCount; i++ {
		startIndex := i * wordBits
		endIndex := startIndex + wordBits
		if endIndex >= len(bin) {
			str = bin[startIndex:]
		} else {
			str = bin[startIndex:endIndex]
		}
		asInt, err := strconv.ParseInt(str, 2, 64)
		if err != nil {
			return "", err
		}
		groups[i] = int(asInt)
	}

	en, err := dictionary()
	if err != nil {
		return "", err
	}
	words := make([]string, wordCount)
	for i, wordIndex := range groups {
		words[i] = en[wordIndex]
	}

	m.sentence = strings.Join(words, " ")

	return m.sentence, nil
}

//GetSeed ...
func (m *Mnemonic) GetSeed() (seed string, e error) {

	sentence, err := m.GetSentence()
	if err != nil {
		e = err
		return
	}
	s := NewSeed(sentence, m.passphrase)
	seed = hex.EncodeToString(s)
	return
}

//NewSeed ...
func NewSeed(mnecmonic, passphrase string) []byte {
	return pbkdf2.Key([]byte(mnecmonic), []byte("mnemonic"+passphrase), 2048, 64, sha512.New)
}

func generateEntropy(bitsCount int) (ent []byte, err error) {
	if bitsCount < minEnt || bitsCount > maxEnt || bitsCount%multiple != 0 {
		err = fmt.Errorf(
			"entropy must between %v-%v and be divisible by %v",
			minEnt, maxEnt, multiple)
		return
	}
	bytesCount := bitsCount / bitsInByte
	ent = make([]byte, bytesCount)
	_, err = rand.Read(ent)
	return
}

/*checksumEntropy A checksum is generated by taking the first
ENT / 32 bits of its SHA256 hash.*/
func checksumEntropy(ent []byte) ([]byte, error) {
	hash := sha256.New()
	_, err := hash.Write(ent)
	if err != nil {
		return nil, err
	}
	sum := hash.Sum(nil)

	bits := len(ent) * bitsInByte
	cs := bits / multiple

	return sum[:cs], nil
}

func splitEntropyToNumbers(ENT []byte) ([]int, error) {
	return []int{}, nil
}

var dict map[string][]string

func dictionary() ([]string, error) {
	if dict == nil {
		dict = make(map[string][]string, 1)
	}
	lang := "english"
	res, ok := dict[lang]
	if ok {
		return res, nil
	}

	size := int(math.Pow(2, wordBits))

	dict[lang] = make([]string, size)

	file, err := os.Open(lang + ".txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		dict[lang][i] = scanner.Text()
		i++
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if i != size {
		log.Fatalf("incomplete dictionary %v, exp lines %v, got %v",
			lang, i, size)
	}

	return dict[lang], nil

}