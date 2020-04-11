/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

func splitStringByStar(str string) [][]byte {
	parts := make([][]byte, 0, 2)
	bytes := trim([]byte(str))

	if len(bytes) > 0 {
		contentBegin := 0

		for contentBegin < len(bytes) {
			// this is correkt: starBegin 0, then part ""
			starBegin := seekStar(bytes, contentBegin)
			part := bytes[contentBegin:starBegin]
			contentBegin = seekContent(bytes, starBegin)
			parts = append(parts, part)
		}
	}
	return parts
}

func trim(bytes []byte) []byte {
	begin := 0
	end := 0

	for i, b := range bytes {
		if b > 32 {
			begin = i
			break
		}
	}
	for i := len(bytes) - 1; i > 0; i-- {
		b := bytes[i]

		if b > 32 {
			end = i + 1
			break
		}
	}
	return bytes[begin:end]
}

func seekStar(bytes []byte, offset int) int {
	for offset < len(bytes) {
		if bytes[offset] == '*' {
			break
		}
		offset++
	}
	return offset
}

func seekContent(bytes []byte, offset int) int {
	for offset < len(bytes) {
		if bytes[offset] != '*' {
			break
		}
		offset++
	}
	return offset
}
