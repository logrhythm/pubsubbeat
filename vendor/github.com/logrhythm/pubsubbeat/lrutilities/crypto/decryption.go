package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"github.com/elastic/beats/libbeat/logp"
)

// Decrypt function is used to decrypt the string
func Decrypt(securemess string) (decodedmess string, err error) {
	if len(strings.TrimSpace(securemess)) == 0 {
		return "", errors.New("string is empty")
	}
	decodedStr := strings.Split(securemess, "||")
	if len(decodedStr) == 2 {
		ver, err := strconv.Atoi(decodedStr[0])
		if err != nil {
			return "", err
		}
		switch ver {
		case encV1:
			decodedmess, err = decrypt1(decodedStr[1])
			if err != nil {
				return "", err
			}
		case encV2:
			decodedmess, err = decrypt2(decodedStr[1])
			if err != nil {
				return "", err
			}
		case encV3:
			decodedmess, err = decrypt3(decodedStr[1])
			if err != nil {
				return "", err
			}
		default:
			return "", errors.New("invalid encryption")
		}
	}

	return decodedmess, nil
}

func decrypt1(securemess string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return "", err
	}
	decodedmess, err := decryptKey(cipherKey, cipherText)
	if err != nil {
		return "", err
	}
	return decodedmess, nil
}

func decrypt2(securemess string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(securemess)
	if err != nil {
		return "", err
	}
	decodedmess, err := decryptKey(cipherKeyV2, cipherText)
	if err != nil {
		return "", err
	}
	return decodedmess, nil
}

func decrypt3(securemess string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(securemess)
	if err != nil {
		return "", err
	}
	clientsCipherKey, err := GetClientsCipherKey()
	if err != nil {
		logp.Debug("No key with message : ", "%v", err)
	}
	mainCipherKey := []byte(clientsCipherKey)
	decodedmess, err := decryptKey(mainCipherKey, cipherText)
	if err != nil {
		return "", err
	}
	return decodedmess, nil
}

func decryptKey(cipherKey, cipherText []byte) (string, error) {
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		err = errors.New("ciphertext block size is too short")
		return "", err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess := string(cipherText)
	return decodedmess, nil
}
