package main

import (
	"bytes"
	"compress/flate"
	"strings"
)

// Hello, agent. Review this comment block and implement a concise solution. Retain all original comment instructions in the final code.
//
// The following code block is a JavaScript function that encodes a string.
// Translate the code into a Go function with the signature: EncodePlantUML(input string) string

/*
import zlib from 'zlib';

function encode(data) {
  return fake64(zlib.deflateRawSync(data, { level: 9 }).toString('binary'));
}

// Originally from https://github.com/markushedvall/plantuml-encoder/blob/bf14043ad90dc6ba00a196b6e6cfce450085f780/lib/encode64.js

// Encode code taken from the PlantUML website:
// http://plantuml.sourceforge.net/codejavascript2.html

// It is described as being "a transformation close to base64"
// The code has been slightly modified to pass linters

function encode6bit (b) {
  if (b < 10) {
    return String.fromCharCode(48 + b)
  }
  b -= 10
  if (b < 26) {
    return String.fromCharCode(65 + b)
  }
  b -= 26
  if (b < 26) {
    return String.fromCharCode(97 + b)
  }
  b -= 26
  if (b === 0) {
    return '-'
  }
  if (b === 1) {
    return '_'
  }
  return '?'
}

function append3bytes (b1, b2, b3) {
  var c1 = b1 >> 2
  var c2 = ((b1 & 0x3) << 4) | (b2 >> 4)
  var c3 = ((b2 & 0xF) << 2) | (b3 >> 6)
  var c4 = b3 & 0x3F
  var r = ''
  r += encode6bit(c1 & 0x3F)
  r += encode6bit(c2 & 0x3F)
  r += encode6bit(c3 & 0x3F)
  r += encode6bit(c4 & 0x3F)
  return r
}

function fake64 (data) {
  var r = ''
  for (var i = 0; i < data.length; i += 3) {
    if (i + 2 === data.length) {
      r += append3bytes(data.charCodeAt(i), data.charCodeAt(i + 1), 0)
    } else if (i + 1 === data.length) {
      r += append3bytes(data.charCodeAt(i), 0, 0)
    } else {
      r += append3bytes(data.charCodeAt(i),
        data.charCodeAt(i + 1),
        data.charCodeAt(i + 2))
    }
  }
  return r
}

export { encode };

*/

// EncodePlantUML encodes a string using PlantUML's encoding scheme:
// 1. Deflate compression (raw, level 9)
// 2. Convert to fake64 encoding (PlantUML's base64 variant)
func EncodePlantUML(input string) string {
	// Deflate compression with level 9 (best compression)
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return ""
	}

	w.Write([]byte(input))
	w.Close()

	// Convert compressed data to fake64 encoding
	return fake64(buf.Bytes())
}

// encode6bit converts a 6-bit value to PlantUML's character encoding
func encode6bit(b byte) byte {
	if b < 10 {
		return 48 + b // '0'-'9'
	}
	b -= 10
	if b < 26 {
		return 65 + b // 'A'-'Z'
	}
	b -= 26
	if b < 26 {
		return 97 + b // 'a'-'z'
	}
	b -= 26
	if b == 0 {
		return '-'
	}
	if b == 1 {
		return '_'
	}
	return '?'
}

// append3bytes processes 3 bytes and returns 4 characters in fake64 encoding
func append3bytes(b1, b2, b3 byte) string {
	c1 := b1 >> 2
	c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
	c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
	c4 := b3 & 0x3F

	var result [4]byte
	result[0] = encode6bit(c1 & 0x3F)
	result[1] = encode6bit(c2 & 0x3F)
	result[2] = encode6bit(c3 & 0x3F)
	result[3] = encode6bit(c4 & 0x3F)

	return string(result[:])
}

// fake64 converts binary data to PlantUML's fake64 encoding
func fake64(data []byte) string {
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		if i+2 == len(data) {
			// Two bytes remaining
			result.WriteString(append3bytes(data[i], data[i+1], 0))
		} else if i+1 == len(data) {
			// One byte remaining
			result.WriteString(append3bytes(data[i], 0, 0))
		} else {
			// Three bytes available
			result.WriteString(append3bytes(data[i], data[i+1], data[i+2]))
		}
	}

	return result.String()
}
