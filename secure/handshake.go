package secure

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"net"

	"golang.org/x/crypto/pbkdf2"
)

func SendHandshake(conn net.Conn, password []byte) (cipher.Stream, error) {
	// first packet to be written, need to handshake
	// we will send [iv] + [salt] + [signature]

	header := make([]byte, aes.BlockSize+KeySize+sha256.Size)
	if _, err := rand.Read(header); err != nil {
		return nil, err
	}
	iv, salt := header[:aes.BlockSize], header[aes.BlockSize:aes.BlockSize+KeySize]

	// create session key
	key := pbkdf2.Key(password, salt, KeyIteration, 2*KeySize, sha256.New)
	aeskey, hmackey := key[:KeySize], key[KeySize:]

	// sign iv and salt and put signature in header
	mac := hmac.New(sha256.New, hmackey)
	mac.Write(header[:aes.BlockSize+KeySize])
	signature := mac.Sum(nil)
	copy(header[aes.BlockSize+KeySize:], signature)

	log.Debugf("sending encryption handshake: remote = %v, local = %v, iv = %v, salt = %v, signature = %v, aes = %v, hmac = %v", conn.RemoteAddr(), conn.LocalAddr(), iv, salt, signature, aeskey, hmackey)

	// create cipher
	block, err := aes.NewCipher(aeskey)
	if err != nil {
		return nil, err
	}
	encrypter := cipher.NewCFBEncrypter(block, iv)

	// write the header
	length := 0
	for length < len(header) {
		size, err := conn.Write(header[length:])
		if err != nil {
			return nil, err
		}
		length += size
	}
	return encrypter, nil
}

func ReceiveHandshake(conn net.Conn, password []byte) (cipher.Stream, error) {
	// first need to receive header
	// [iv] + [salt] + [signature]
	header := make([]byte, aes.BlockSize+KeySize+sha256.Size)
	length := 0
	for length < len(header) {
		size, err := conn.Read(header[length:])
		if err != nil {
			return nil, err
		}
		length += size
	}
	iv, salt, signature := header[:aes.BlockSize], header[aes.BlockSize:aes.BlockSize+KeySize], header[aes.BlockSize+KeySize:]

	// create session key
	key := pbkdf2.Key(password, salt, KeyIteration, 2*KeySize, sha256.New)
	aeskey, hmackey := key[:KeySize], key[KeySize:]

	log.Debugf("received encryption handshake: remote = %v, local = %v, iv = %v, salt = %v, signature = %v, aes = %v, hmac = %v", conn.RemoteAddr(), conn.LocalAddr(), iv, salt, signature, aeskey, hmackey)

	// sign iv and salt and validate signature in header
	mac := hmac.New(sha256.New, hmackey)
	mac.Write(header[:aes.BlockSize+KeySize])
	if !bytes.Equal(mac.Sum(nil), signature) {
		return nil, ErrInvalidPassword
	}

	// create cipher
	block, err := aes.NewCipher(aeskey)
	if err != nil {
		return nil, err
	}
	return cipher.NewCFBDecrypter(block, iv), nil
}
