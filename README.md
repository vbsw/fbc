# fbc

[![Go Report Card](https://goreportcard.com/badge/github.com/vbsw/fbc)](https://goreportcard.com/report/github.com/vbsw/fbc) [![Stability: Experimental](https://masterminds.github.io/stability/experimental.svg)](https://masterminds.github.io/stability/experimental.html)

## About
fbc (file by content) allows various commands on files filtered by their content. fbc is published on <https://github.com/vbsw/fbc>.

## Copyright
Copyright 2020, Vitali Baumtrok (vbsw@mailbox.org).

fbc is distributed under the Boost Software License, version 1.0. (See accompanying file LICENSE or copy at http://www.boost.org/LICENSE_1_0.txt)

fbc is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the Boost Software License for more details.

## Usage

	fbc (INFO | ( COMMAND INPUT-DIR {OUTPUT-DIR CONTENT-FILTER} ))

	INFO
		-h, --help        print this help
		-v, --version     print version
		-e, --example     print example
		--copyright       print copyright
	COMMAND
		cp                copy files
		mv                move files
		rm                delete files

## Examples

Copy any file containing the words "bob" and "alice"

	$ fbc cp ./ ../bak bob alice

Move text files containing the words "bob" and "alice"

	$ fbc mv ./*.txt ../bak bob alice

Delete text files containing the words "bob" and "alice"

	$ fbc rm ./*.txt bob alice

## References
- https://golang.org/doc/install
- https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
