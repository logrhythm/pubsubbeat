package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	cipherKey                = []byte("0123456789012345")
	cipherKeyV2              = []byte("CCEF7CFA0DCB2237012FAE9EB09CCD70")
	clientsCipherKeyPath     = "/app/cmd/beats/cipherstore/"
	clientsCipherKeyFileName = "cipher_key.yml"
)

const (
	encV1 = 1
	encV2 = 2
	encV3 = 3
)

//Encrypt function is used to encrypt the string
func Encrypt(message ...string) (encmess string, err error) {
	var mainCipherKey []byte
	var clientsCipherKey, msg string
	msg = message[0]
	if len(message) == 2 {
		clientsCipherKey = message[1]
	} else {
		clientsCipherKey = ""
	}
	if len(strings.TrimSpace(msg)) == 0 {
		return "", errors.New("string is empty")
	}
	plainText := []byte(msg)
	var encVersion int
	if err == nil && clientsCipherKey != "" {
		mainCipherKey = []byte(clientsCipherKey)
		encVersion = encV3
	} else {
		mainCipherKey = cipherKeyV2
		encVersion = encV2
	}

	block, err := aes.NewCipher(mainCipherKey)
	if err != nil {
		return "", err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.StdEncoding.EncodeToString(cipherText)
	finalEnc := fmt.Sprintf("%d%s%s", encVersion, "||", encmess)
	return finalEnc, nil
}

// CipherKeyStruct encapsulates cipher key data
type CipherKeyStruct struct {
	CipherKey string `yaml:"cipher_key"`
}

// GetClientsCipherKey is to get the cipher key of the client if any found
func GetClientsCipherKey() (string, error) {
	path := filepath.Join(clientsCipherKeyPath, url.QueryEscape(clientsCipherKeyFileName))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	var cipherKeyVal CipherKeyStruct
	err = yaml.Unmarshal(data, &cipherKeyVal)
	if err != nil {
		return "", err
	}
	return cipherKeyVal.CipherKey, nil
}
