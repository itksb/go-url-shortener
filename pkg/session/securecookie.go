package session

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
)

type Codec interface {
	Encode(name string, value interface{}) (string, error)
	Decode(name string, value string, dst interface{}) error
}

type SecureCookie struct {
	hashKey   []byte
	blockKey  []byte
	maxLength int
}

// NewSecureCookie creates a new SecureCookie
//
//	hashKey and blockKey are required.
//
// The blockKey size must correspond to the key size of the crypto algorithm
// 2 * aes.BlockSize in our case
func NewSecureCookie(hashKey, blockKey []byte) (*SecureCookie, error) {
	if len(hashKey) == 0 {
		return nil, errors.New("hashKey cannot have length 0")
	}
	if len(blockKey) == 0 {
		return nil, errors.New("blockKey cannot have length 0")
	}
	if len(blockKey) != 2*aes.BlockSize {
		return nil, errors.New("blockKey size MUST have 2 * aes.BlockSize = 32 length")
	}

	return &SecureCookie{
		hashKey:   hashKey,
		blockKey:  blockKey,
		maxLength: 4096,
	}, nil

}

// Encode encodes a cookie value.
//
// It serializes, encrypts, signs with a message authentication code,
// and finally encodes the value.
// ToDo: add timestamp for verification of max-age
func (sc *SecureCookie) Encode(name string, value interface{}) (string, error) {
	var err error
	// Serialize
	binaryData, err := sc.serialize(value)
	if err != nil {
		return "", err
	}
	// Encrypt
	binaryData, err = sc.encrypt(binaryData)
	if err != nil {
		return "", err
	}
	base64Encoded := make([]byte, base64.URLEncoding.EncodedLen(len(binaryData)))
	base64.URLEncoding.Encode(base64Encoded, binaryData)
	binaryData = base64Encoded

	// Inspired by Gorilla secure cookie
	// Create MAC - message authentication code for "name|value"
	binaryData = []byte(fmt.Sprintf("%s|%s|", name, binaryData))
	authCode := sc.createAuthenticationCode(binaryData[:len(binaryData)-1])
	// Append authCode, remove name.
	binaryData = append(binaryData, authCode...)[len(name)+1:]
	// Encode to base64
	base64Encoded = make([]byte, base64.URLEncoding.EncodedLen(len(binaryData)))
	base64.URLEncoding.Encode(base64Encoded, binaryData)
	binaryData = base64Encoded

	// Check length (длина куки не может быть больше 4кБ)
	if len(binaryData) > sc.maxLength {
		return "", errors.New("the length of the generated value is too long")
	}

	return string(binaryData), nil
}

// Decode - decodes, verifies, decrypts and deserializes a cookie value
func (sc *SecureCookie) Decode(name string, value string, dst interface{}) error {
	// check the len
	if len(value) > sc.maxLength {
		return errors.New("the length of the input value is too long")
	}
	// decode base64
	binaryData, err := decodeBase64([]byte(value))
	if err != nil {
		return err
	}
	// verify sign - message authentication code. Value is "value|authCode".
	split := bytes.SplitN(binaryData, []byte("|"), 2)
	if len(split) != 2 {
		return errors.New("invalid sign (message authentication code)")
	}
	valueReceived := split[0]
	authCodeReceived := split[1]
	binaryData = append([]byte(name+"|"), binaryData[:len(binaryData)-len(authCodeReceived)-1]...)
	err = sc.verifyAuthenticationCode(binaryData, authCodeReceived)
	if err != nil {
		return err
	}

	// Decrypt
	binaryData, err = decodeBase64(valueReceived)
	if err != nil {
		return err
	}

	binaryData, err = sc.decrypt(binaryData)
	if err != nil {
		return err
	}

	// Deserialize
	err = sc.deserialize(binaryData, dst)
	if err != nil {
		return err
	}

	return nil
}

func decodeBase64(value []byte) ([]byte, error) {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		return nil, errors.New("base64 decoding error")
	}
	return decoded[:b], nil

}

// Serialize - always use gob as serializer, as simplest solution
func (sc *SecureCookie) serialize(src interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(src)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Deserialize - always use gob decoder.
// dst - should be a pointer
func (sc *SecureCookie) deserialize(src []byte, dst interface{}) error {
	buffer := bytes.NewBuffer(src)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(dst)
	return err
}

func (sc *SecureCookie) encrypt(src []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(sc.blockKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	//Для работы алгоритма GCM нужно дополнительно сгенерировать
	//вектор инициализации из 12 байт.
	//Вектор должен быть уникальным для каждой процедуры шифрования.
	//Если переиспользовать один и тот же вектор, можно атаковать алгоритм,
	//подавая на вход данные с разницей в один байт, и по косвенным
	//признакам вычислить ключ шифрования.

	// создаём вектор инициализации
	nonce, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		return nil, err
	}

	// Encrypt
	encryptedValue := aesgcm.Seal(nil, nonce, src, nil)

	// Where to store Nonce (вектор инициализации):
	//https://crypto.stackexchange.com/questions/57895/would-it-be-safe-to-store-gcm-nonce-in-the-encrypted-output
	// nonce is usual 12 bytes, so not a problem to extract it
	return append(nonce, encryptedValue...), nil
}

func (sc *SecureCookie) decrypt(value []byte) ([]byte, error) {
	aesblock, err := aes.NewCipher(sc.blockKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(value) < nonceSize {
		return nil, errors.New("wrong value size")
	}
	nonce := value[:nonceSize]
	cipherValues := value[nonceSize:]

	decrypted, err := aesgcm.Open(nil, nonce, cipherValues, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (sc *SecureCookie) createAuthenticationCode(value []byte) []byte {
	// подписываем алгоритмом HMAC, используя SHA256
	h := hmac.New(sha256.New, sc.hashKey)
	h.Write(value)
	dst := h.Sum(nil)
	return dst
}

// verifyAuthenticationCode - verifies message authentication code (sign)
// value - decoded data message
// authCodeToValidate - sign of the value
func (sc *SecureCookie) verifyAuthenticationCode(value []byte, authCodeToValidate []byte) error {
	authCode2 := sc.createAuthenticationCode(value)
	if hmac.Equal(authCode2, authCodeToValidate) {
		return nil
	}
	return errors.New("message authentication code is wrong")
}
