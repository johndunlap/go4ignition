package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
)

// FromFile Load the specified file into memory and return its contents as a string
func FromFile(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		println("Error reading file: " + err.Error())
	}
	return string(content)
}

// DirExists Return true if a directory exists at the given path and false otherwise
func DirExists(path string) bool {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		return false
	}

	return info.IsDir()
}

// FileExists Return true if a file exists with the given name and false otherwise
func FileExists(filename string) bool {
	_, err := os.Stat(filename)

	if err != nil {
		return false
	}
	return !os.IsNotExist(err)
}

// CreateDir Attempt to create a directory with the given path
func CreateDir(path string) {
	err := os.Mkdir(path, 0700)

	if err != nil {
		println("Error creating directory: " + err.Error())
		return
	}
}

// ToFile Write the specified string to the specified file, overwriting its contents
func ToFile(filename string, text string) {
	err := os.WriteFile(filename, []byte(text), 0600)

	if err != nil {
		println("Error writing file: " + err.Error())
		return
	}
}

// Md5ToPath Convert the given hash and data directory into a fully qualified path
func Md5ToPath(hash string, dataDir string) string {
	chars := []rune(hash)
	path := ""

	for x := 0; x < 32; x += 2 {
		if x > 1 {
			path += "/"
		}

		path += string(chars[x]) + string(chars[x+1])
	}

	return dataDir + "/" + path
}

// Md5sum Compute the MD5 hash of the given string
func Md5sum(text string) string {
	// Create a new MD5 hash object
	hash := md5.New()

	// Write the input string to the hash
	hash.Write([]byte(text))

	// Get the resulting hash as a byte slice
	hashBytes := hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	return hex.EncodeToString(hashBytes)
}

// Check Panic if there's something to panic about
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// FileHandler Used as an argument to MkTemp
type FileHandler func(*os.File)

// MkTemp Handling the temporary file with a Lambda allows the temporary file to be automatically deleted once it is
// no longer needed
func MkTemp(fn FileHandler) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "tmp")
	Check(err)

	// Try to clean up after ourselves
	defer os.Remove(tmpFile.Name())

	// Do something with the temporary file
	fn(tmpFile)
}

// Obfuscate Output Go code which embeds the given string in the go executable in an obfuscated manner
func Obfuscate(input string) string {
	output := "package main\nimport \"fmt\"\n\n\n"

	finalFunc := "func final() string {\n\t return "

	for x := 0; x < len(input); x++ {
		funcName := "c" + strconv.Itoa(rand.IntN(1000000000))
		output += "func " + funcName + "() string {\n"

		chr := string(input[x])
		if chr == "\n" {
			chr = "\\n"
		}

		output += "\treturn \"" + chr + "\"\n"
		output += "}\n\n"
		finalFunc += "\t" + funcName + "()"

		if len(input)-1 != x {
			finalFunc += " +"
		}
		finalFunc += "\n"
	}

	finalFunc += "}\n"
	output += finalFunc

	return output
}

func Md5(content string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}
